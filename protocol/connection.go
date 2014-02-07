package protocol

import (
	"net"
	"sync"
	"gobs/biller"
	"fmt"
	"log"
	"strings"
)

type Connection struct {
	Conn net.Conn
	Quit chan struct{}
	WaitGroup *sync.WaitGroup
	Biller *biller.Biller
	Zone string
}

func (c *Connection) Send(msg fmt.Stringer) (int, error) {
	log.Print("Send: " + msg.String())
	return c.Conn.Write([]byte(msg.String() + "\n"))
}

func (c *Connection) Disconnect() {
	c.Conn.Close()
}

func (c *Connection) Listen() {
	c.WaitGroup.Add(1)
	defer c.WaitGroup.Done()
	defer c.Disconnect()

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

func (c *Connection) listen(out chan string) {
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

func (c *Connection) handleMessage(buffer string) {
	fields := strings.Split(buffer, ":")
	if len(fields) == 0 {
		return
	}

	// Check if client tries to do something without connecting first
	if c.Zone != "" && fields[0] != "CONNECT" {
		log.Printf("Not yet connected: %s\n", fields)
		c.Send(ConnectBad{"Not yet connected"})
		c.Disconnect()
		return
	}

	if fields[0] == "CONNECT" {
		msg, err := ParseConnect(fields)
		if err != nil {
			log.Printf("Could not parse connection message: %s\n", err)
			c.Send(ConnectBad{"Wrong connection message"})
			c.Disconnect()
			return
		}

		if c.Zone != "" {
			log.Printf("Already connected: %s\n", msg.Zonename)
			c.Send(ConnectBad{"Already connected"})
			c.Disconnect()
			return
		}

		if msg.Version.Major != PROTOCOL_MAJOR || msg.Version.Minor != PROTOCOL_MINOR {
			c.Send(ConnectBad{"Protocol mismatch"})
			c.Disconnect()
		} else {
			if msg.Zonename == "" {
				c.Send(ConnectBad{"Invalid zonename"})
				c.Disconnect()
			}

			// TODO: Check if zone already exists

			// TODO: Check if create zone was successfull
			c.Biller.CreateZone(msg.Zonename)
			c.Zone = msg.Zonename
			c.Send(ConnectOk{})
		}
	} else if fields[0] == "PLOGIN" {
		msg, err := ParsePlogin(fields)
		if err != nil {
			log.Printf("Could not parse plogin message: %s", err)
			return
		}

		c.handlePlogin(msg)
	} else if fields[0] == "PENTERARENA" {
		msg, err := ParsePenterArena(fields)
		if err != nil {
			log.Printf("Could not parse penterarena message: %s", err)
			return
		}

		c.handlePenterArena(msg)
	}
}
