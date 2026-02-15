# pool-go

Go client library for the [POOL](https://github.com/amosdavis/POOL) (Protected Orchestrated Overlay Link) protocol.

## Overview

`pool-go` provides two layers:

| Package | Purpose |
|---------|---------|
| `poolioc` | Low-level ioctl wrapper for `/dev/pool` — mirrors every pool.h struct and constant |
| `pool` | High-level `net.Conn` / `net.Listener` API for idiomatic Go networking |

## Requirements

- Linux with the `pool.ko` kernel module loaded
- Go 1.21+
- `/dev/pool` character device accessible

## Installation

```bash
go get github.com/amosdavis/pool-go
```

## Quick Start

### Echo Server

```go
package main

import (
    "io"
    "log"

    "github.com/amosdavis/pool-go/pool"
)

func main() {
    ln, err := pool.Listen("pool", ":9253")
    if err != nil {
        log.Fatal(err)
    }
    defer ln.Close()

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Print(err)
            continue
        }
        go func() {
            defer conn.Close()
            io.Copy(conn, conn)
        }()
    }
}
```

### Client

```go
package main

import (
    "fmt"
    "log"

    "github.com/amosdavis/pool-go/pool"
)

func main() {
    conn, err := pool.Dial("pool", "10.0.0.1:9253")
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    conn.Write([]byte("hello POOL"))

    buf := make([]byte, 4096)
    n, _ := conn.Read(buf)
    fmt.Println(string(buf[:n]))
}
```

## API Reference

### High-Level (`pool` package)

```go
// Listen and accept connections
ln, err := pool.Listen("pool", ":9253")
conn, err := ln.Accept()

// Dial a peer (IPv4, IPv6, or hostname)
conn, err := pool.Dial("pool", "10.0.0.1:9253")
conn, err := pool.Dial("pool6", "[::1]:9253")
conn, err := pool.DialTimeout("pool", "host.example.com:9253", 5*time.Second)

// net.Conn interface
n, err := conn.Read(buf)
n, err := conn.Write(data)
conn.SetDeadline(time.Now().Add(10 * time.Second))
conn.Close()

// Session telemetry
telem, err := conn.Telemetry()
fmt.Printf("RTT: %dμs, Loss: %d%%\n", telem.RttUs, telem.LossPercent)

// Multi-channel I/O
ch5, err := conn.OpenChannel(5)
ch5.Write([]byte("channel 5 data"))
ch5.Close()
```

### Low-Level (`poolioc` package)

```go
// Open the device
dev, err := poolioc.Open()
defer dev.Close()

// Listen / Connect
dev.Listen(9253)
idx, err := dev.Connect(poolioc.ConnectReq{...})

// Send / Receive
dev.SendBytes(idx, 0, []byte("hello"))
n, err := dev.RecvBytes(idx, 0, buf)

// Sessions
sessions, err := dev.Sessions()
dev.CloseSession(idx)

// Channels
dev.ChannelSubscribe(idx, 5)
bitmap, err := dev.ChannelList(idx)
```

## Address Formats

| Network | Format | Example |
|---------|--------|---------|
| `pool`  | IPv4 or IPv6, auto-detect | `10.0.0.1:9253`, `[::1]:9253` |
| `pool4` | IPv4 only | `10.0.0.1:9253` |
| `pool6` | IPv6 only | `[::1]:9253`, `[2001:db8::1]:9253` |

## Errors

| Error | Meaning |
|-------|---------|
| `pool.ErrSessionFull` | Kernel session table full |
| `pool.ErrAuthFailed` | Handshake authentication failed |
| `pool.ErrClosed` | Connection already closed |
| `pool.ErrTimeout` | Deadline exceeded |
| `pool.ErrMessageTooLarge` | Payload exceeds MaxPayload |
| `pool.ErrNetUnreachable` | Peer unreachable |

## Examples

See the [`examples/`](examples/) directory:

- **[echo](examples/echo/)** — Echo server and interactive client
- **[filetransfer](examples/filetransfer/)** — Send/receive files over POOL
- **[telemetry](examples/telemetry/)** — Live session telemetry monitoring

## Testing

BDD tests use [godog](https://github.com/cucumber/godog) (Cucumber for Go):

```bash
cd tests
go test -v ./steps/
```

> **Note:** Tests require a running POOL kernel module. Integration tests
> that need a peer are marked as `Pending` and can be wired up in a CI
> environment with two POOL nodes.

## Architecture

```
Application
    │
    ▼
┌──────────────────┐
│   pool package   │  net.Conn, net.Listener, Dial, Listen
├──────────────────┤
│ poolioc package  │  Open, ioctl, raw structs
├──────────────────┤
│   /dev/pool      │  Character device (kernel)
├──────────────────┤
│    pool.ko       │  Kernel module
└──────────────────┘
```

## License

MIT — see [LICENSE](LICENSE).
