package main

import (
	"log"
	"net"
	"os"
	"os/signal"
)

var clients map[int]net.Conn

func socketHandler(fd net.Conn, clientN int) {
	for {
		buf := make([]byte, 512)
		n, err := fd.Read(buf)
		if err != nil {
			log.Println(err)
			fd.Close()
			delete(clients, clientN)
			return
		}

		data := buf[0:n]
		log.Println(string(data))
	}
}

func listenForClients(l net.Listener) {
	var client int
	for {
		fd, err := l.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		clients[client] = fd
		go socketHandler(fd, client)
		client++
	}
}

func main() {
	clients = make(map[int]net.Conn)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	l, err := net.Listen("unix", "/tmp/led_cloud")
	if err != nil {
		log.Fatalln(err)
	}

	go listenForClients(l)

	<-c

	for _, c := range clients {
		if err := c.Close(); err != nil {
			log.Println(err)
		}
	}

	if err := l.Close(); err != nil {
		log.Println(err)
	}
	os.Exit(0)

}
