# Plan: Go Library for POOL (`pool-go`)

## Problem

There is no Go library for writing applications that use the POOL protocol. Developers who want to build POOL-native applications in Go have no SDK. The POOL kernel module exposes an ioctl API via `/dev/pool`, but Go applications need an idiomatic wrapper.

## Approach

Create a new Git repo at `C:\Users\adavi068\pool-go` (module path `github.com/amosdavis/pool-go`) with two layers:

1. **Low-level `poolioc` package** — thin Go wrapper around every POOL ioctl, mirroring pool.h structs as Go types with `unsafe.Sizeof`-compatible layouts. Opens `/dev/pool`, calls `ioctl()` via `syscall.Syscall`.

2. **High-level `pool` package** — idiomatic Go networking API implementing `net.Conn`, `net.Listener`, `io.Reader`, `io.Writer`, `io.Closer`. Provides `pool.Dial()`, `pool.Listen()`, `listener.Accept()`. Handles address parsing (IPv4, IPv6, hostnames), session lifecycle, channel multiplexing, telemetry access, and graceful close.

BDD tests in Cucumber/Go (godog) for both layers.

## Package Structure

```
pool-go/
├── go.mod                          # github.com/amosdavis/pool-go
├── go.sum
├── README.md                       # Usage guide with examples
├── LICENSE                         # MIT
├── poolioc/                        # Low-level ioctl wrapper
│   ├── poolioc.go                  # Device open/close, ioctl helper
│   ├── types.go                    # Go equivalents of pool.h structs/constants
│   ├── connect.go                  # Connect, Listen, Stop
│   ├── session.go                  # Sessions, CloseSess
│   ├── data.go                     # Send, Recv
│   ├── channel.go                  # Channel subscribe/unsubscribe/list
│   └── doc.go                      # Package documentation
├── pool/                           # High-level net.Conn API
│   ├── conn.go                     # Conn implementing net.Conn
│   ├── listener.go                 # Listener implementing net.Listener
│   ├── dial.go                     # Dial, DialTimeout
│   ├── addr.go                     # Addr implementing net.Addr
│   ├── session.go                  # Session info, telemetry access
│   ├── channel.go                  # Multi-channel conn support
│   ├── errors.go                   # Typed errors
│   └── doc.go                      # Package documentation
├── examples/                       # Example applications
│   ├── echo/                       # Echo server/client
│   │   └── main.go
│   ├── filetransfer/               # File transfer demo
│   │   └── main.go
│   └── telemetry/                  # Telemetry monitoring
│       └── main.go
└── tests/                          # BDD tests
    ├── features/
    │   ├── poolioc.feature         # Low-level ioctl tests
    │   └── pool.feature            # High-level net.Conn tests
    └── steps/
        ├── godog_test.go
        ├── poolioc_steps.go
        └── pool_steps.go
```

## Todos

### Group 1: Repository Setup
- `repo-init` — Create repo, go.mod, LICENSE, .gitignore

### Group 2: Low-Level `poolioc` Package
- `ioc-types` — Define Go struct equivalents for all pool.h types (ConnectReq, SendReq, RecvReq, SessionInfo, SessionList, ChannelReq, Header, Telemetry, Address) and constants (ioctl numbers, packet types, flags, states, error codes, limits)
- `ioc-device` — Device type wrapping /dev/pool fd with Open/Close/ioctl helper
- `ioc-connect` — Listen, Connect, Stop methods
- `ioc-session` — Sessions (list), CloseSession methods
- `ioc-data` — Send, Recv methods
- `ioc-channel` — ChannelSubscribe, ChannelUnsubscribe, ChannelList methods

### Group 3: High-Level `pool` Package
- `hl-addr` — Addr type implementing net.Addr (network="pool", string="[::1]:9253")
- `hl-errors` — Typed error types (ErrSessionFull, ErrAuthFailed, ErrDecryptFailed, etc.)
- `hl-conn` — Conn type implementing net.Conn (Read, Write, Close, LocalAddr, RemoteAddr, SetDeadline, SetReadDeadline, SetWriteDeadline)
- `hl-listener` — Listener type implementing net.Listener (Accept, Close, Addr); Dial and DialTimeout functions
- `hl-session` — SessionInfo and Telemetry accessors on Conn
- `hl-channel` — ChannelConn for reading/writing on specific channels

### Group 4: Examples
- `ex-echo` — Echo server/client example
- `ex-filetransfer` — File transfer example
- `ex-telemetry` — Telemetry monitoring example

### Group 5: Tests & Documentation
- `bdd-tests` — BDD feature files and step definitions for both packages
- `readme` — README.md with installation, quick start, API overview, examples

## Dependencies

```
repo-init → (none)
ioc-types → repo-init
ioc-device → ioc-types
ioc-connect → ioc-device
ioc-session → ioc-device
ioc-data → ioc-device
ioc-channel → ioc-device
hl-addr → ioc-types
hl-errors → repo-init
hl-conn → ioc-data, ioc-session, hl-addr, hl-errors
hl-listener → ioc-connect, hl-conn, hl-addr
hl-session → ioc-session, hl-conn
hl-channel → ioc-channel, hl-conn
ex-echo → hl-listener, hl-conn
ex-filetransfer → hl-listener, hl-conn
ex-telemetry → hl-session
bdd-tests → hl-listener, hl-conn, ioc-device
readme → ex-echo, bdd-tests
```

## Key Design Decisions

1. **Linux-only build constraint** — `poolioc` uses `//go:build linux` since it requires `/dev/pool` and Linux ioctl syscalls.

2. **Struct layout** — Go structs in `poolioc/types.go` must match the C struct byte layout exactly (use `[16]byte` for peer_addr, explicit padding fields). Verified with `unsafe.Sizeof`.

3. **Conn.Read/Write** — Internally use `poolioc.Send`/`poolioc.Recv` with the default channel (0). For multi-channel, use `ChannelConn`.

4. **Address parsing** — `pool.Dial("pool", "::1:9253")` and `pool.Dial("pool4", "10.0.0.1:9253")` and `pool.Dial("pool6", "[::1]:9253")`. Uses `net.SplitHostPort` + `net.ResolveIPAddr`.

5. **Deadline support** — `SetDeadline`/`SetReadDeadline`/`SetWriteDeadline` use goroutine-based timeouts around blocking ioctls (the kernel API is synchronous).

6. **Thread safety** — `Conn` methods are safe for concurrent use (multiple goroutines can Read/Write). Internal mutex protects device fd and session state.

7. **Telemetry** — `conn.Telemetry()` returns the latest `pool_telemetry` struct (RTT, jitter, loss, throughput, MTU, etc.) via `POOL_IOC_SESSIONS`.

## Notes

- The Go library does NOT re-implement the POOL protocol — it wraps the kernel module's ioctl interface. The kernel module must be loaded (`pool.ko`).
- `poolioc` is intentionally 1:1 with pool.h — if pool.h changes, only `poolioc/types.go` needs updating.
- The high-level `pool` package never exposes ioctl details — users work with `net.Conn` and `net.Listener`.
- IPv6 is fully supported (the kernel module already handles dual-stack).
