package sockets

import (
	"bufio"
	"log"
	"logplus"
	"net"
	"os"
)

type Client struct{}

func (c *Client) Start(network, address string) {
	conn, err := net.Dial(network, address)
	logplus.LogIfErrorFatal(err)
	defer conn.Close()

	buffer := make(buffer, 1<<12)
	scanner := bufio.NewScanner(os.Stdin)
	for {
		buffer.Zero()
		scanner.Scan()
		logplus.LogIfErrorFatal(scanner.Err())
		buffer = []byte(scanner.Text())
		_, err := conn.Write(buffer)
		logplus.LogIfErrorFatal(err)
	}
}

func (c *Client) OneShotWrite(network, address string, content []byte) {
	conn, err := net.Dial(network, address)
	logplus.LogIfErrorFatal(err)
	defer conn.Close()

	//will recieve a byte
	buffer := make(buffer, 1)

	_, err = conn.Write(content)
	logplus.LogIfErrorFatal(err)
	_, err = conn.Read(buffer)
	logplus.LogIfErrorFatal(err)
	if buffer[0] == byte(0) {
		log.Println("Signal", string(content), "doesn't exist!")
		return
	}
	log.Println("Signal", string(content), "sent!")
}
