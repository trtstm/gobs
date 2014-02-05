package protocol

import (
	"log"
	"fmt"
	"net"
	"sync"
	"strings"
	"gobs/biller"
//	"time"
)

const (
	SERVER_NAME string = "Gobs 0.0"
	SERVER_NETWORK = "" // Eg. SSC
	PROTOCOL_MAJOR uint = 1
	PROTOCOL_MINOR uint = 3
	PROTOCOL_PATCH uint = 1
)

type Client struct {
	Conn net.Conn
	Quit chan struct{}
	WaitGroup *sync.WaitGroup
	Biller *biller.Biller
}

func (c *Client) Send(msg fmt.Stringer) (int, error) {
	log.Print("Send: " + msg.String())
	return c.Conn.Write([]byte(msg.String() + "\n"))
}

func (c *Client) Listen() {
	c.WaitGroup.Add(1)
	defer c.WaitGroup.Done()
	defer c.Conn.Close()

	bufCh := make(chan string)
	go c.listen(bufCh)

	for {
		select {
			case <- c.Quit:
				return

			case buffer, ok := <- bufCh:
				if !ok {
					return
				}

				c.handleMessage(buffer)
		}
	}
}

func (c *Client) listen(out chan string) {
	c.WaitGroup.Add(1)
	defer c.WaitGroup.Done()

	buf := make([]byte, 1023)
	for {
		n, err := c.Conn.Read(buf)
		if err != nil || n < 1 || n > 1023 {
			close(out)
			return
		}

		log.Print("Recv: " + string(buf[:n]))
		out <- string(buf)
	}
}

func (c *Client) handleMessage(buffer string) {
	fields := strings.Split(buffer, ":")
	if len(fields) == 0 {
		return
	}

	if fields[0] == "CONNECT" {
		msg, err := ParseConnect(fields)
		if err != nil {
			log.Printf("Could not parse connection message: %s", err)
			return
		}

		if msg.Version.Major != PROTOCOL_MAJOR || msg.Version.Minor != PROTOCOL_MINOR {
			c.Send(ConnectBad{"Protocol mismatch"})
		} else {
			c.Send(ConnectOk{})
		}
	} else if fields[0] == "PLOGIN" {
		msg, err := ParsePlogin(fields)
		if err != nil {
			log.Printf("Could not parse plogin message: %s", err)
			return
		}

		// TODO: Check user credentials etc...

		//answer := protocol.Pok{msg.Pid, "should be mpty", msg.Name, "some squad", 123, 60, "1-2-1999 6:13:35"}
		//answer := protocol.Pbad{msg.Pid, true, "User not existing noob :p"}
		//client.Send(answer)
		c.handlePlogin(msg)
	}

	fmt.Fprintf(c.Conn, "PING\n")
}

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
	Pid uint
	Rtext string
	Name string
	Squad string
	BillerId uint
	Usage uint
	FirstUsed string
}

func (p Pok) String() string {
	return fmt.Sprintf("POK:%d:%s:%s:%s:%d:%d:%s", p.Pid, p.Rtext, p.Name, p.Squad, p.BillerId, p.Usage, p.FirstUsed)
}

type Pbad struct {
	Pid uint // Pid that game server gave us
	NewName bool
	Rtext string
}

func (p Pbad) String() string {
	userExists := "0"
	if p.NewName {
		userExists = "1"
	}

	return fmt.Sprintf("PBAD:%d:%s:%s", p.Pid, userExists, p.Rtext)
}
