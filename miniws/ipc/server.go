package ipc

import (
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

type Server struct{}

func (s *Server) Start(recvBind func(string, []string) bool, network, address string) int {
	socket, err := net.Listen(network, address)
	LogIfErrorFatal(err)

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
		LogIfErrorFatal(err)

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
				LogIfErrorFatal(err)
				fullstring := string(buffer)
				arguments := strings.Split(fullstring, " ")
				ret := recvBind(arguments[0], arguments[1:])
				if ret {
					conn.Write([]byte{1})
				} else {
					conn.Write([]byte{0})
				}
				buffer.Zero()
			}
		}(conn)
	}
}
