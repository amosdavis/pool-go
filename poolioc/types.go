//go:build linux

package poolioc

import "unsafe"

// --------------------------------------------------------------------
// Protocol constants
// --------------------------------------------------------------------

const (
	Version       = 1
	VersionPQC    = 2 // hybrid X25519 + ML-KEM-768
	HeaderSize    = 80
	HMACSize      = 32 // SHA-256
	SessionIDSize = 16 // 128-bit
	KeySize       = 32 // X25519 / ChaCha20
	NonceSize     = 12 // ChaCha20-Poly1305
	TagSize       = 16 // Poly1305
	AddrSize      = 32 // 256-bit POOL address
	MaxPayload    = 65535
	ListenBacklog = 128
	DefaultMTU    = 1400
	MinMTU        = 512
	HeartbeatSec  = 5
	RekeyPackets  = 1 << 28 // 256M packets
	RekeySec      = 3600
	MaxSessions   = 64
	MaxChannels   = 256
	ListenPort    = 9253
	PuzzleDiff    = 8
	MaxFrags      = 256
	FragTimeoutMS = 5000
	IPProto       = 253 // IANA experimental
)

// Transport modes.
const (
	TransportTCP  = 0
	TransportRaw  = 1
	TransportAuto = 2
)

// --------------------------------------------------------------------
// Packet types (4-bit field)
// --------------------------------------------------------------------

const (
	PktInit      = 0x0
	PktChallenge = 0x1
	PktResponse  = 0x2
	PktData      = 0x3
	PktAck       = 0x4
	PktHeartbeat = 0x5
	PktRekey     = 0x6
	PktClose     = 0x7
	PktConfig    = 0x8
	PktRollback  = 0x9
	PktDiscover  = 0xA
	PktJournal   = 0xB
)

// --------------------------------------------------------------------
// Flags (16-bit)
// --------------------------------------------------------------------

const (
	FlagEncrypted   = 1 << 0
	FlagCompressed  = 1 << 1
	FlagPriority    = 1 << 2
	FlagFragment    = 1 << 3
	FlagLastFrag    = 1 << 4
	FlagRequireAck  = 1 << 5
	FlagTelemetry   = 1 << 6
	FlagRollbackRdy = 1 << 7
	FlagConfigLock  = 1 << 8
	FlagJournalSync = 1 << 9
)

// --------------------------------------------------------------------
// Session states
// --------------------------------------------------------------------

const (
	StateIdle        = 0
	StateInitSent    = 1
	StateChallenged  = 2
	StateEstablished = 3
	StateRekeying    = 4
	StateClosing     = 5
)

// --------------------------------------------------------------------
// POOL error codes (wire protocol)
// --------------------------------------------------------------------

const (
	ErrAuthFail        = 0x01
	ErrDecryptFail     = 0x02
	ErrSeqInvalid      = 0x03
	ErrFragTimeout     = 0x04
	ErrMTUExceeded     = 0x05
	ErrConfigReject    = 0x06
	ErrRekeyFail       = 0x07
	ErrJournalFull     = 0x08
	ErrOverload        = 0x09
	ErrVersionMismatch = 0x0A
)

// --------------------------------------------------------------------
// Journal change types
// --------------------------------------------------------------------

const (
	JournalConnect    = 1
	JournalDisconnect = 2
	JournalConfig     = 3
	JournalRekey      = 4
	JournalError      = 5
	JournalData       = 6
)

// --------------------------------------------------------------------
// Channel operations
// --------------------------------------------------------------------

const (
	ChanSubscribe   = 1
	ChanUnsubscribe = 2
	ChanList        = 3
)

// --------------------------------------------------------------------
// Ioctl numbers
//
// _IOC encoding: dir(2) | size(14) | type(8) | nr(8)
// POOL_IOC_MAGIC = 'P' = 0x50
// --------------------------------------------------------------------

const iocMagic = 'P'

// iocEncode computes the Linux ioctl number.
func iocEncode(dir, nr, size uintptr) uintptr {
	return (dir << 30) | (size << 16) | (uintptr(iocMagic) << 8) | nr
}

const (
	iocWrite    = 1
	iocRead     = 2
	iocReadWrite = iocRead | iocWrite
)

var (
	iocListen    = iocEncode(iocWrite, 1, unsafe.Sizeof(uint16(0)))
	iocConnect   = iocEncode(iocWrite, 2, unsafe.Sizeof(ConnectReq{}))
	iocSend      = iocEncode(iocWrite, 3, unsafe.Sizeof(SendReq{}))
	iocRecv      = iocEncode(iocReadWrite, 4, unsafe.Sizeof(RecvReq{}))
	iocSessions  = iocEncode(iocReadWrite, 5, unsafe.Sizeof(SessionList{}))
	iocCloseSess = iocEncode(iocWrite, 6, unsafe.Sizeof(uint32(0)))
	iocStop      = iocEncode(0, 7, 0)
	iocChannel   = iocEncode(iocReadWrite, 8, unsafe.Sizeof(ChannelReq{}))

	// Exported aliases for inspection/testing.
	IocListen  = iocListen
	IocConnect = iocConnect
)

// --------------------------------------------------------------------
// Ioctl request/response structs — must match pool.h byte layout
// --------------------------------------------------------------------

// ConnectReq matches struct pool_connect_req.
type ConnectReq struct {
	PeerAddr   [16]byte // IPv4-mapped (::ffff:x.x.x.x) or native IPv6
	PeerPort   uint16
	AddrFamily uint8 // syscall.AF_INET or syscall.AF_INET6
	_reserved  [5]byte
}

// SendReq matches struct pool_send_req.
type SendReq struct {
	SessionIdx uint32
	Channel    uint8
	Flags      uint8
	_reserved  uint16
	Len        uint32
	DataPtr    uint64 // userspace pointer
}

// RecvReq matches struct pool_recv_req.
type RecvReq struct {
	SessionIdx uint32
	Channel    uint8
	Flags      uint8
	_reserved  uint16
	Len        uint32 // in: buffer size, out: bytes received
	DataPtr    uint64 // userspace pointer
}

// Telemetry matches struct pool_telemetry.
type Telemetry struct {
	RTTNs         uint64
	JitterNs      uint64
	LossRatePPM   uint32
	ThroughputBps uint32
	MTUCurrent    uint16
	QueueDepth    uint16
	UptimeNs      uint64
	RekeyCount    uint32
	ConfigVersion uint32
}

// SessionInfo matches struct pool_session_info.
type SessionInfo struct {
	Index      uint32
	PeerAddr   [16]byte
	PeerPort   uint16
	AddrFamily uint8
	State      uint8
	SessionID  [SessionIDSize]byte
	BytesSent  uint64
	BytesRecv  uint64
	PacketsSent uint64
	PacketsRecv uint64
	RekeyCount uint32
	Telem      Telemetry
}

// SessionList matches struct pool_session_list.
type SessionList struct {
	Count       uint32
	MaxSessions uint32
	InfoPtr     uint64 // userspace pointer to []SessionInfo
}

// ChannelReq matches struct pool_channel_req.
type ChannelReq struct {
	SessionIdx uint32
	Channel    uint8
	Operation  uint8 // ChanSubscribe, ChanUnsubscribe, ChanList
	_reserved  uint16
	Result     uint32
	DataPtr    uint64 // for LIST: pointer to [256]byte bitmap
}

// Header matches struct pool_header (80 bytes, on-wire).
type Header struct {
	VerType    uint8
	_reserved0 uint8
	Flags      uint16
	Seq        uint64
	Ack        uint64
	SessionID  [SessionIDSize]byte
	Timestamp  uint64
	PayloadLen uint16
	Channel    uint8
	_reserved1 uint8
	HMAC       [HMACSize]byte
}

// Address matches struct pool_address (256-bit).
type Address struct {
	TypeVersion uint32
	OrgID       uint64
	SegmentID   uint64
	NodeID      uint64
	Checksum    uint32
}

// FragHeader matches struct pool_frag_header.
type FragHeader struct {
	MsgID      uint32
	FragOffset uint16
	TotalLen   uint16
}

// JournalEntry matches struct pool_journal_entry (variable length).
type JournalEntry struct {
	Timestamp      uint64
	ConfigVerBefore uint32
	ConfigVerAfter  uint32
	ChangeHash     [32]byte
	ChangeType     uint16
	DetailLength   uint16
}

// --------------------------------------------------------------------
// IPv4-mapped IPv6 helpers
// --------------------------------------------------------------------

// IPv4ToMapped converts a host-byte-order IPv4 address to an
// IPv4-mapped IPv6 address (::ffff:x.x.x.x).
func IPv4ToMapped(ip4 uint32) [16]byte {
	var addr [16]byte
	addr[10] = 0xFF
	addr[11] = 0xFF
	addr[12] = byte(ip4 >> 24)
	addr[13] = byte(ip4 >> 16)
	addr[14] = byte(ip4 >> 8)
	addr[15] = byte(ip4)
	return addr
}

// MappedToIPv4 extracts the host-byte-order IPv4 address from an
// IPv4-mapped IPv6 address.
func MappedToIPv4(addr [16]byte) uint32 {
	return uint32(addr[12])<<24 | uint32(addr[13])<<16 |
		uint32(addr[14])<<8 | uint32(addr[15])
}

// IsV4Mapped returns true if addr is an IPv4-mapped IPv6 address.
func IsV4Mapped(addr [16]byte) bool {
	for i := 0; i < 10; i++ {
		if addr[i] != 0 {
			return false
		}
	}
	return addr[10] == 0xFF && addr[11] == 0xFF
}
