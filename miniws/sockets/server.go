package sockets

import (
	"io"
	"logplus"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Server struct{}

func (s *Server) Start(RecvBind func(string, []string) byte, network, address string) int {
	socket, err := net.Listen(network, address)
	logplus.LogIfErrorFatal(err)

	//Cleanup the socket file
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		os.Remove(address)
		os.Exit(1)
	}()

	for { //Infinite loop

		//Accept connection
		conn, err := socket.Accept()
		logplus.LogIfErrorFatal(err)

		//Handle the connection
		//in a separate goroutine
		go func(conn net.Conn) {
			defer conn.Close()
			//Create a buffer (slice)
			//for incoming data
			buffer := make(buffer, 1<<12)

			for {
				//Read data
				_, err := conn.Read(buffer)
				if err == io.EOF {
					conn.Close()
					break
				}
				logplus.LogIfErrorFatal(err)
				fullstring := string(buffer)
				arguments := strings.Split(fullstring, " ")
				ret := RecvBind(arguments[0], arguments[1:])
				conn.Write([]byte{ret})
				buffer.Zero()
			}
		}(conn)
	}
}
