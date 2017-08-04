package collectors

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/iwondory/agent_manager/event"
	"github.com/iwondory/agent_manager/libs"
)

const (
	msgBufSize = 1024
)

var (
	iv  = []byte("2981eeca66b5c3cd")                 // internal vector
	key = []byte("c43ac86d84469030f28c0a9656b1c533") // key
	fs  = []byte("|")                                // field separator
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
			if err != nil {
				log.Printf("Read error: " + err.Error())
				continue
			}

			data_enc := buf[:n]
			data_dec, err := libs.Decrypt(key, iv, data_enc)
			if err != nil {
				log.Printf("Decryption error: " + err.Error())
				continue
			}

			agent, err := parse(data_dec)
			if err != nil {
				log.Printf("Parse error: " + err.Error())
				continue
			}
			spew.Dump(agent)
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
	spew.Dump(cols)
	if len(cols) != 9 {
		return nil, fmt.Errorf("Invalid columns")
	}

	agent := event.Agent{
		Guid:               string(cols[1]),
		OsVersionNumber:    libs.ByteToFloat64(cols[4]),
		OsBit:              libs.ByteToInt64(cols[6]),
		OsIsServer:         libs.ByteToInt64(cols[5]),
		ComputerName:       string(cols[3]),
		Eth:                string(cols[2]),
		FullPolicyVersion:  string(cols[7]),
		TodayPolicyVersion: string(cols[8]),
		Rdate:              time.Now(),
		Udate:              time.Now(),
	}

	return &agent, nil
}
