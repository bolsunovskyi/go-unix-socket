package main

import (
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/jgarff/rpi_ws281x/golang/ws2811"

	"strconv"
)

var clients map[int]net.Conn

const (
	pin        = 18
	count      = 16
	brightness = 255
)

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

		data := string(buf[0:n])
		log.Println(data)
		udata, err := strconv.ParseUint(data, 10, 32)
		if err != nil {
			log.Println(err)
			continue
		}

		colorWipe(uint32(udata))
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
	defer ws2811.Fini()
	if err := ws2811.Init(pin, count, brightness); err != nil {
		log.Fatalln(err)
	}

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

func colorWipe(color uint32) error {
	for i := 0; i < count; i++ {
		ws2811.SetLed(i, color)
		err := ws2811.Render()
		if err != nil {
			ws2811.Clear()
			return err
		}
	}

	return nil
}
