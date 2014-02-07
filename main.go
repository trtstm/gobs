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
	Connection *protocol.Connection
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

			connection := protocol.Connection{Conn: conn, Quit: quit, WaitGroup: &waitGroup, Biller: biller}
			go connection.Listen()
		}
	}()

	<- quit
	listener.Close()

	waitGroup.Wait()
}
