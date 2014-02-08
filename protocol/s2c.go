package protocol

import (
	//	"log"
	"fmt"
	//	"time"
)

const (
	SERVER_NAME    string = "Gobs 0.0"
	SERVER_NETWORK        = "" // Eg. SSC
	PROTOCOL_MAJOR uint   = 1
	PROTOCOL_MINOR uint   = 3
	PROTOCOL_PATCH uint   = 1
)

type ConnectOk struct {
}

func (c ConnectOk) String() string {
	return fmt.Sprintf("CONNECTOK:%s:%s", SERVER_NAME, SERVER_NETWORK)
}

type ConnectBad struct {
	Reason string
}

func (c ConnectBad) String() string {
	return fmt.Sprintf("CONNECTBAD:%s:%s:%s", SERVER_NAME, SERVER_NETWORK, c.Reason)
}

type Pok struct {
	Pid       uint
	Rtext     string
	Name      string
	Squad     string
	BillerId  uint
	Usage     uint
	FirstUsed string
}

func (p Pok) String() string {
	return fmt.Sprintf("POK:%d:%s:%s:%s:%d:%d:%s", p.Pid, p.Rtext, p.Name, p.Squad, p.BillerId, p.Usage, p.FirstUsed)
}

type Pbad struct {
	Pid     uint // Pid that game server gave us
	NewName bool
	Rtext   string
}

func (p Pbad) String() string {
	userExists := "0"
	if p.NewName {
		userExists = "1"
	}

	return fmt.Sprintf("PBAD:%d:%s:%s", p.Pid, userExists, p.Rtext)
}
