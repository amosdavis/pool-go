//go:build linux

// Command filetransfer sends or receives a file over POOL.
//
// Usage:
//
//	filetransfer -listen :9253 -out received.dat   # receiver
//	filetransfer -connect 10.0.0.1:9253 -in data.bin  # sender
package main

import (
	"flag"
	"io"
	"log"
	"os"

	"github.com/amosdavis/pool-go/pool"
)

func main() {
	listen := flag.String("listen", "", "listen address (receiver)")
	connect := flag.String("connect", "", "connect address (sender)")
	inFile := flag.String("in", "", "input file to send")
	outFile := flag.String("out", "", "output file to receive")
	flag.Parse()

	switch {
	case *listen != "" && *outFile != "":
		runReceiver(*listen, *outFile)
	case *connect != "" && *inFile != "":
		runSender(*connect, *inFile)
	default:
		flag.Usage()
		os.Exit(1)
	}
}

func runReceiver(addr, outPath string) {
	ln, err := pool.Listen("pool", addr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	defer ln.Close()
	log.Printf("waiting for sender on %s", ln.Addr())

	conn, err := ln.Accept()
	if err != nil {
		log.Fatalf("accept: %v", err)
	}
	defer conn.Close()
	log.Printf("receiving from %s", conn.RemoteAddr())

	f, err := os.Create(outPath)
	if err != nil {
		log.Fatalf("create %s: %v", outPath, err)
	}
	defer f.Close()

	n, err := io.Copy(f, conn)
	if err != nil {
		log.Printf("receive error: %v", err)
	}
	log.Printf("received %d bytes → %s", n, outPath)
}

func runSender(addr, inPath string) {
	f, err := os.Open(inPath)
	if err != nil {
		log.Fatalf("open %s: %v", inPath, err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		log.Fatalf("stat: %v", err)
	}

	conn, err := pool.Dial("pool", addr)
	if err != nil {
		log.Fatalf("dial: %v", err)
	}
	defer conn.Close()
	log.Printf("sending %s (%d bytes) to %s", inPath, info.Size(), conn.RemoteAddr())

	buf := make([]byte, 4096)
	var total int64
	for {
		n, readErr := f.Read(buf)
		if n > 0 {
			if _, writeErr := conn.Write(buf[:n]); writeErr != nil {
				log.Fatalf("write: %v", writeErr)
			}
			total += int64(n)
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			log.Fatalf("read: %v", readErr)
		}
	}
	log.Printf("sent %d bytes", total)
}
