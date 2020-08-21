package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

const (
	CONN_HOST = "0.0.0.0"
	CONN_PORT = "9999"
	CONN_TYPE = "tcp"
)

var (
	connections []*Client
)

func main() {
	// Listen for incoming connections.
	l, err := net.Listen(CONN_TYPE, CONN_HOST+":"+CONN_PORT)

	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()

	fmt.Println("Listening on " + CONN_HOST + ":" + CONN_PORT)

	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Save connection
		connection := &Client{
			conn: conn,
		}

		connection.init()

		connections = append(connections, connection)
		// Handle connections in a new goroutine.
		go handleRequest(connection)
	}
}

// Handles incoming requests.
func handleRequest(conn *Client) {
	for {
		var msg, _ = conn.Recv()
		switch t := msg.(type) {
		case *Message:
			broadcast(t, t.To)
		case *lnetRequest:
			broadcast(t, t.To)
		case *Data:
			broadcast(t, t.To)
		}
	}
}

func removeConn(conn *Client) {
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
		if strings.EqualFold(connection.nickname, name) {
			err := connection.Send(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}
