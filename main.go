package main

import (
	"expvar"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/iwondory/agent_manager/collectors"
	"github.com/iwondory/agent_manager/engine"
)

const (
	DefaultDataDir         = "./temp"
	DefaultBatchSize       = 1000
	DefaultBatchDuration   = 5000
	DefaultBatchMaxPending = 1000
	DefaultUDPPort         = "localhost:19902"
	DefaultInputFormat     = "syslog"
	DefaultMonitorIface    = "localhost:8080"
)

var (
	stats = expvar.NewMap("server")
	fs    *flag.FlagSet
)

func main() {
	// Set CPU count
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
	)

	// Start engine
	duration := time.Duration(*batchDuration) * time.Millisecond
	batcher := engine.NewBatcher(duration, *batchSize, *batchMaxPending, *datadir)
	errChan := make(chan error)
	batcher.Start(errChan)
	go logDrain("error", errChan)

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
		log.Println("signal received, shutting down...")
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
