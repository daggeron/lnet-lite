package cmd

import (
	"flag"
	"fmt"
	"github.com/daggeron/lnet-lite/cmd/lnet"
	"io"
	"log"
	"net"
	"os"
	"strings"
)



var (
	server   = flag.String("server", "localhost:9999", "server")
	flagDebug   = flag.Bool("debug", false, "enable debug")
)

var (
	connections []*lnet.Connection
)



func Main() {
	// Listen for incoming connections.
	l, err := net.Listen("tcp", *server)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	fmt.Printf("Listening on %s\n", *server)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Save connection
		connection := lnet.New(conn, *flagDebug)

		connections = append(connections, connection)

		go handleRequest(connection)
	}
}

// Handles incoming requests.
func handleRequest(conn *lnet.Connection) {
	for {

		msg, err := conn.Recv()

		if err != nil {
			if err == io.EOF {
				removeConn(conn)
				conn.Close()
				return
			}
			log.Println(err)
			return
		}

		switch t := msg.(type) {
		case *lnet.Message:
			broadcast(t, t.To)
		case *lnet.LNETRequest:
			broadcast(t, t.To)
		case *lnet.Data:
			broadcast(t, t.To)
		}

	}
}

func removeConn(conn *lnet.Connection) {
	var i int
	for i = range connections {
		if connections[i] == conn {
			break
		}
	}
	connections = append(connections[:i], connections[i+1:]...)
}

func broadcast(msg interface{}, name string) {
	for _, connection := range connections {
		if strings.EqualFold(connection.NickName, name) {
			err := connection.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
