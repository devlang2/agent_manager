package collectors

import (
	"bytes"
	"errors"
	"expvar"
	"log"
	"net"
	"strconv"
	"time"

	"github.com/devlang2/agent_manager/event"
	"github.com/devlang2/golibs/encryption"
)

const (
	msgBufSize = 1024
)

var (
	iv    = []byte("2981eeca66b5c3cd")                 // internal vector
	key   = []byte("c43ac86d84469030f28c0a9656b1c533") // key
	fs    = []byte("|")                                // field separator
	stats = expvar.NewMap("udp")
)

type UDPCollector struct {
	format string
	addr   *net.UDPAddr
}

func (s *UDPCollector) Start(c chan<- *event.Agent) error {
	conn, err := net.ListenUDP("udp", s.addr)
	if err != nil {
		return err
	}

	go func() {
		buf := make([]byte, msgBufSize)
		for {
			n, addr, err := conn.ReadFromUDP(buf)
			//			spew.Dump(buf[:n])
			if err != nil {
				log.Printf("Read error: " + err.Error())
				continue
			}

			data_enc := append(iv, buf[:n]...)
			data_dec, err := encryption.Decrypt(key, data_enc)
			if err != nil {
				log.Printf("Decryption error: " + err.Error())
				continue
			}

			agent, err := parse(data_dec)
			if err != nil {
				log.Printf("Parse error: " + err.Error())
				log.Printf("Data: ", string(buf[:n]))
				stats.Add("parseFailed", 1)
				continue
			}
			agent.IP = addr.IP
			c <- agent
		}
	}()
	return nil
}

func (s *UDPCollector) Addr() net.Addr {
	return s.addr
}

func parse(b []byte) (*event.Agent, error) {
	cols := bytes.Split(b, fs)

	if len(cols) != 9 {
		return nil, errors.New("Invalid columns")
	}

	agent := event.Agent{
		Guid:               string(cols[1]),
		OsVersionNumber:    ByteToFloat64(cols[4]),
		OsBit:              ByteToInt64(cols[6]),
		OsIsServer:         ByteToInt64(cols[5]),
		ComputerName:       string(cols[3]),
		Eth:                string(cols[2]),
		FullPolicyVersion:  string(cols[7]),
		TodayPolicyVersion: string(cols[8]),
		Rdate:              time.Now(),
		Udate:              time.Now(),
	}

	return &agent, nil
}

func ByteToFloat64(b []byte) float64 {
	f, _ := strconv.ParseFloat(string(b), 64)
	return f
}

func ByteToInt64(b []byte) int64 {
	str := string(b)
	i, _ := strconv.ParseInt(str, 10, 64)
	return i
}
