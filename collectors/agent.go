package collectors

import (
	"net"

	"github.com/iwondory/udpserver/event"
)

type AgentCollector struct {
	format string
	addr   *net.UDPAddr
}

func (this *AgentCollector) Start(c chan<- *event.Event) error {
	return nil
}

func (this *AgentCollector) Addr() net.Addr {
	return this.addr
}
