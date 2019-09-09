package main

import (
	"bufio"
	"flag"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func main() {
	// Parse arguments
	port := flag.Int("port", 8000, "Port to accept connections on.")
	flag.Parse()

	// Start server
	l, err := net.Listen("tcp", ":"+strconv.Itoa(*port))
	if err != nil {
		log.Panicln(err)
	}

	// Accept connections
	log.Println("Kuma TCP Echo - Listening to connections on port", strconv.Itoa(*port))
	defer l.Close()
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Panicln(err)
		}
		go handleRequest(conn)
	}
}

func handleRequest(conn net.Conn) {
	defer conn.Close()

	for {
		buf := make([]byte, 1024)
		size, err := conn.Read(buf)
		if err != nil {
			return
		}
		data := buf[:size]
		reader := bufio.NewReader(strings.NewReader(string(data)))
		logReq, err := http.ReadRequest(reader)
		if err == nil {
			log.Printf("%s connection: %s %s", logReq.Proto, logReq.Method, logReq.URL)
			conn.Write(data)
			conn.Close()
		} else {
			log.Printf("TCP connection: %v", data)
			conn.Write(data)
		}

	}
}
