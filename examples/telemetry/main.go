//go:build linux

// Command telemetry connects to a POOL peer and continuously prints
// session telemetry (RTT, jitter, loss, throughput).
//
// Usage:
//
//	telemetry -connect 10.0.0.1:9253
//	telemetry -listen :9253
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/amosdavis/pool-go/pool"
)

func main() {
	listen := flag.String("listen", "", "listen address")
	connect := flag.String("connect", "", "connect address")
	interval := flag.Duration("interval", 2*time.Second, "polling interval")
	flag.Parse()

	var conn *pool.Conn
	var err error

	switch {
	case *connect != "":
		conn, err = pool.Dial("pool", *connect)
		if err != nil {
			log.Fatalf("dial: %v", err)
		}
	case *listen != "":
		ln, lErr := pool.Listen("pool", *listen)
		if lErr != nil {
			log.Fatalf("listen: %v", lErr)
		}
		log.Printf("waiting on %s", ln.Addr())
		c, aErr := ln.Accept()
		if aErr != nil {
			log.Fatalf("accept: %v", aErr)
		}
		conn = c.(*pool.Conn)
	default:
		flag.Usage()
		os.Exit(1)
	}

	defer conn.Close()
	log.Printf("monitoring session %d to %s", conn.SessionIndex(), conn.RemoteAddr())

	for {
		t, tErr := conn.Telemetry()
		if tErr != nil {
			log.Printf("telemetry: %v", tErr)
			return
		}

		state, _ := conn.SessionState()
		fmt.Printf("[%s] state=%s rtt=%dns jitter=%dns loss=%dppm throughput=%d B/s mtu=%d queue=%d\n",
			time.Now().Format("15:04:05"),
			state,
			t.RTTNs,
			t.JitterNs,
			t.LossRatePPM,
			t.ThroughputBps,
			t.MTUCurrent,
			t.QueueDepth,
		)

		time.Sleep(*interval)
	}
}
