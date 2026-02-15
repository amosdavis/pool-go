//go:build linux

// Command echo is a simple POOL echo server and client.
//
// Usage:
//
//	echo -listen :9253           # start server
//	echo -connect 10.0.0.1:9253  # connect and type messages
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/amosdavis/pool-go/pool"
)

func main() {
	listen := flag.String("listen", "", "address to listen on (e.g. :9253)")
	connect := flag.String("connect", "", "address to connect to (e.g. 10.0.0.1:9253)")
	flag.Parse()

	switch {
	case *listen != "":
		runServer(*listen)
	case *connect != "":
		runClient(*connect)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func runServer(addr string) {
	ln, err := pool.Listen("pool", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	log.Printf("listening on %s", ln.Addr())

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleEcho(conn)
	}
}

func handleEcho(conn net.Conn) {
	defer conn.Close()
	log.Printf("new session from %s", conn.RemoteAddr().String())

	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("read: %v", err)
			return
		}
		if _, err := conn.Write(buf[:n]); err != nil {
			log.Printf("write: %v", err)
			return
		}
	}
}

func runClient(addr string) {
	conn, err := pool.Dial("pool", addr)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	log.Printf("connected to %s", conn.RemoteAddr())

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if err != nil {
				log.Printf("read: %v", err)
				return
			}
			fmt.Printf("echo: %s\n", buf[:n])
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		if _, err := conn.Write(scanner.Bytes()); err != nil {
			log.Fatalf("write: %v", err)
		}
	}
}
