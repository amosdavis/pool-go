// Package pool provides idiomatic Go networking over the POOL protocol.
//
// It implements [net.Conn] and [net.Listener] on top of the POOL kernel
// module, allowing applications to use POOL with the same patterns as
// standard Go networking code.
//
// Quick start:
//
//	// Server
//	ln, _ := pool.Listen("pool", ":9253")
//	conn, _ := ln.Accept()
//	io.Copy(conn, conn) // echo
//
//	// Client
//	conn, _ := pool.Dial("pool", "10.0.0.1:9253")
//	conn.Write([]byte("hello"))
//
// The "pool" network accepts IPv4 addresses, IPv6 addresses (including
// bracket notation), and hostnames. All traffic is encrypted with
// ChaCha20-Poly1305 and authenticated with HMAC-SHA256.
//
// This package requires Linux with the pool.ko kernel module loaded.
package pool
