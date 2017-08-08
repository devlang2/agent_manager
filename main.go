package main

import (
	"encoding/json"
	"expvar"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/devlang2/agent_manager/collectors"
	"github.com/devlang2/agent_manager/engine"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	DefaultDataDir         = "./temp"
	DefaultBatchSize       = 2000
	DefaultBatchDuration   = 5000
	DefaultBatchMaxPending = 10000
	DefaultUDPPort         = "localhost:19902"
	DefaultInputFormat     = "syslog"
	DefaultMonitorIface    = "localhost:8080"
)

var (
	stats = expvar.NewMap("server")
	fs    *flag.FlagSet
)

type Config struct {
	Database struct {
		Host     string `json:"host"`
		Username string `json:"username"`
		Database string `json:"database"`
		Port     string `json:"port"`
		Password string `json:"password"`
	} `json:"database"`
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// Flag set
	fs = flag.NewFlagSet("", flag.ExitOnError)
	var (
		batchSize       = fs.Int("batchsize", DefaultBatchSize, "Indexing batch size")
		batchDuration   = fs.Int("duration", DefaultBatchDuration, "Indexing batch timeout, in milliseconds")
		batchMaxPending = fs.Int("maxpending", DefaultBatchMaxPending, "Maximum pending index events")
		udpIface        = fs.String("udp", DefaultUDPPort, "Syslog server UDP bind address in the form host:port")
		inputFormat     = fs.String("input", DefaultInputFormat, "Message format of input (only syslog supported)")
		datadir         = fs.String("datadir", DefaultDataDir, "Set data directory")
		monitorIface    = fs.String("monitor", DefaultMonitorIface, "TCP Bind address for monitoring server in the form host:port.")
		debug           = fs.Bool("debug", false, "Debug mode")
	)
	fs.Usage = printHelp
	fs.Parse(os.Args[1:])

	// Load configuration
	config, err := loadConfig("config.json")
	if err != nil {
		panic(err.Error())
	}

	// Set log output
	mw := io.MultiWriter(os.Stderr, &lumberjack.Logger{
		Filename:   "server.log",
		MaxSize:    1, // MB
		MaxBackups: 3,
		MaxAge:     1, //days
	})
	log.SetOutput(mw)

	log.Printf("Starting server.. (batchSize: %d, batchDuration: %dms, batchMaxPending: %d)\n", *batchSize, *batchDuration, *batchMaxPending)

	// Start engine
	duration := time.Duration(*batchDuration) * time.Millisecond
	batcher := engine.NewBatcher(duration, *batchSize, *batchMaxPending, *datadir)
	errChan := make(chan error)
	batcher.Start(errChan, debug)
	go logDrain("error", errChan)

	// Connect to database
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8&allowAllFiles=true", config.Database.Username, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.Database)
	engine.InitDatabase(connStr)

	// Start UDP collector
	if err := startUDPCollector(*udpIface, *inputFormat, batcher); err != nil {
		log.Fatalf("failed to start UDP collector: %s", err.Error())
	}
	log.Printf("UDP collector listening to %s", *udpIface)

	// Start monitoring
	startStatusMonitoring(monitorIface)

	// Stop
	waitForSignals()
}

func startStatusMonitoring(monitorIface *string) error {
	http.HandleFunc("/e", func(w http.ResponseWriter, r *http.Request) {
	})

	go http.ListenAndServe(*monitorIface, nil)
	return nil
}

func logDrain(msg string, errChan <-chan error) {
	for {
		select {
		case err := <-errChan:
			if err != nil {
				log.Printf("%s: %s", msg, err.Error())
			}
		}
	}
}

func waitForSignals() {
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt, syscall.SIGTERM)
	select {
	case <-signalCh:
		log.Println("Stopping server..")
	}
}

func startUDPCollector(iface, format string, batcher *engine.Batcher) error {
	collector, err := collectors.NewCollector("udp", iface, format, nil)
	if err != nil {
		return fmt.Errorf("failed to create UDP collector: %s", err.Error())
	}
	if err := collector.Start(batcher.C()); err != nil {
		return fmt.Errorf("failed to start UDP collector: %s", err.Error())
	}

	return nil
}

func printHelp() {
	fmt.Println("amserver [options]")
	fs.PrintDefaults()
}

func loadConfig(file string) (Config, error) {
	var config Config
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		return config, err
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config, nil
}
