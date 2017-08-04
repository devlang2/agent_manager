package event

import (
	"net"
	"time"
)

type Agent struct {
	Guid               string
	Mac                string
	IP                 net.IP
	OsVersionNumber    float64
	OsBit              int64
	OsIsServer         int64
	ComputerName       string
	Eth                string
	FullPolicyVersion  string
	TodayPolicyVersion string
	Rdate              time.Time
	Udate              time.Time
	LastInspectionDate time.Time
}

func NewAgent() *Agent {
	return &Agent{Rdate: time.Now()}
}
