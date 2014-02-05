package main

import (
	"net"
	"log"
	"gobs/protocol"
	"gobs/biller"	
	"os"
	"os/signal"
	"sync"
)

type ClientInfo struct {
	Client *protocol.Client
	HasQuit chan bool
}

func interruptHandler(quit chan struct{}) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	close(quit)
}

func main() {
	listener, err := net.Listen("tcp", ":1850")
	if err != nil {
		log.Fatalln("Could not start listening: ", err.Error())
	}

	quit := make(chan struct{}, 1)
	waitGroup := sync.WaitGroup{}
	biller := biller.NewBiller("biller.db")
	if biller == nil {
		log.Fatalln("Could not create biller")
	}

	go interruptHandler(quit)

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				break
			}

			client := protocol.Client{Conn: conn, Quit: quit, WaitGroup: &waitGroup, Biller: biller}
			go client.Listen()
		}
	}()

	<- quit
	listener.Close()

	waitGroup.Wait()
}

func connectMsgHandler(client *protocol.Client, fields []string) {
	msg, err := protocol.ParseConnect(fields)
	if err != nil {
		log.Printf("Could not parse connection message: %s", err)
		return
	}

	if msg.Version.Major != protocol.PROTOCOL_MAJOR || msg.Version.Minor != protocol.PROTOCOL_MINOR {
		client.Send(protocol.ConnectBad{"Protocol mismatch"})
	} else {
		client.Send(protocol.ConnectOk{})
	}
}

func ploginHandler(client *protocol.Client, fields []string) {
	msg, err := protocol.ParsePlogin(fields)
	if err != nil {
		log.Printf("Could not parse plogin message: %s", err)
		return
	}

	// TODO: Check user credentials etc...

	//answer := protocol.Pok{msg.Pid, "should be mpty", msg.Name, "some squad", 123, 60, "1-2-1999 6:13:35"}
	answer := protocol.Pbad{msg.Pid, true, "User not existing noob :p"}
	client.Send(answer)
}
