// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package generators

import (
	"encoding/binary"
	"math/rand"
	"net"
	"time"
)

// GenerateSFlowPacket generates a minimal valid sFlow v5 packet
// Structure based on sFlow v5 specification
func GenerateSFlowPacket() []byte {
	buf := make([]byte, 0, 256)

	// sFlow v5 header
	buf = binary.BigEndian.AppendUint32(buf, 5)           // Version 5
	buf = binary.BigEndian.AppendUint32(buf, 1)           // IP version (1 = IPv4)
	buf = append(buf, net.ParseIP("10.0.0.1").To4()...)   // Agent IP
	buf = binary.BigEndian.AppendUint32(buf, 1)           // Sub-agent ID
	buf = binary.BigEndian.AppendUint32(buf, rand.Uint32()) // Sequence number
	buf = binary.BigEndian.AppendUint32(buf, uint32(time.Now().Unix())) // Uptime (ms)
	buf = binary.BigEndian.AppendUint32(buf, 1)           // Sample count

	// Flow sample header
	// Sample length = 32 bytes (flow sample fields) + 8 bytes (record header) + 60 bytes (record content) = 100
	buf = binary.BigEndian.AppendUint32(buf, 1)   // Sample type: flow sample
	buf = binary.BigEndian.AppendUint32(buf, 100) // Sample length

	// Flow sample data
	buf = binary.BigEndian.AppendUint32(buf, rand.Uint32()) // Sequence number
	buf = binary.BigEndian.AppendUint32(buf, 1)             // Source ID type
	buf = binary.BigEndian.AppendUint32(buf, 100)           // Sampling rate
	buf = binary.BigEndian.AppendUint32(buf, rand.Uint32()) // Sample pool
	buf = binary.BigEndian.AppendUint32(buf, 0)             // Drops
	buf = binary.BigEndian.AppendUint32(buf, 1)             // Input interface
	buf = binary.BigEndian.AppendUint32(buf, 2)             // Output interface
	buf = binary.BigEndian.AppendUint32(buf, 1)             // Flow record count

	// Flow record: Raw packet header
	// Record length = 16 bytes (SampledHeader fields) + 42 bytes (header data) = 58
	// SampledHeader fields: Protocol(4) + FrameLength(4) + Stripped(4) + HeaderLength(4) = 16
	// Header data: Ethernet(14) + IP(20) + TCP ports(8) = 42
	buf = binary.BigEndian.AppendUint32(buf, 1)  // Record type: raw packet header (FLOW_TYPE_RAW)
	buf = binary.BigEndian.AppendUint32(buf, 60) // Record length (padded to 4-byte boundary)

	// Raw packet header format (SampledHeader fields)
	buf = binary.BigEndian.AppendUint32(buf, 1)    // Header protocol (1 = Ethernet)
	buf = binary.BigEndian.AppendUint32(buf, 1500) // Frame length
	buf = binary.BigEndian.AppendUint32(buf, 0)    // Stripped bytes
	buf = binary.BigEndian.AppendUint32(buf, 42)   // Header data length (Eth + IP + TCP ports)

	// Ethernet + IP + TCP header (42 bytes total)
	// Ethernet header (14 bytes)
	buf = append(buf, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55) // Dst MAC
	buf = append(buf, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb) // Src MAC
	buf = append(buf, 0x08, 0x00)                          // EtherType: IPv4

	// IP header (20 bytes)
	buf = append(buf, 0x45, 0x00)                                         // Version/IHL, DSCP
	buf = binary.BigEndian.AppendUint16(buf, 1480)                        // Total length
	buf = binary.BigEndian.AppendUint16(buf, uint16(rand.Uint32()&0xFFFF)) // ID
	buf = append(buf, 0x00, 0x00)                                         // Flags/Fragment
	buf = append(buf, 64, 6)                                              // TTL, Protocol (TCP)
	buf = append(buf, 0x00, 0x00)                                         // Checksum (0 for fake)

	// Random source/dest IPs
	srcIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	dstIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	buf = append(buf, srcIP...)
	buf = append(buf, dstIP...)

	// TCP/UDP ports (8 bytes for extraction)
	srcPort := uint16(rand.Intn(65535-1024) + 1024) // Random high port
	dstPort := SamplePorts[rand.Intn(len(SamplePorts))]
	buf = binary.BigEndian.AppendUint16(buf, srcPort) // Source port
	buf = binary.BigEndian.AppendUint16(buf, dstPort) // Destination port
	buf = append(buf, 0x00, 0x00, 0x00, 0x00)         // TCP sequence number (4 bytes padding to 42)

	// Padding to reach record length of 60 (need 2 more bytes)
	buf = append(buf, 0x00, 0x00)

	return buf
}

// GenerateIPFIXPacket generates a valid IPFIX packet with common flow fields
// Structure based on RFC 7011
// Fields: protocolIdentifier, sourceTransportPort, sourceIPv4Address, destinationTransportPort,
//         destinationIPv4Address, octetDeltaCount, packetDeltaCount
func GenerateIPFIXPacket() []byte {
	buf := make([]byte, 0, 256)

	// Calculate sizes:
	// Template: 4 (header) + 4 (template header) + 7*4 (7 field specs) = 36 bytes
	// Data: 4 (header) + 1 (proto) + 2 (srcPort) + 4 (srcIP) + 2 (dstPort) + 4 (dstIP) + 8 (bytes) + 8 (packets) = 33 bytes
	// Total: 16 (header) + 36 (template set) + 33 (data set) = 85 bytes

	// IPFIX Message Header (16 bytes)
	buf = binary.BigEndian.AppendUint16(buf, 10)                        // Version 10 (IPFIX)
	buf = binary.BigEndian.AppendUint16(buf, 85)                        // Length
	buf = binary.BigEndian.AppendUint32(buf, uint32(time.Now().Unix())) // Export time
	buf = binary.BigEndian.AppendUint32(buf, rand.Uint32())             // Sequence number
	buf = binary.BigEndian.AppendUint32(buf, 12345)                     // Observation domain ID

	// Template Set Header (4 bytes) + Template Record (32 bytes) = 36 bytes
	buf = binary.BigEndian.AppendUint16(buf, 2)  // Set ID = 2 (Template Set)
	buf = binary.BigEndian.AppendUint16(buf, 36) // Set length

	// Template Record Header
	buf = binary.BigEndian.AppendUint16(buf, 256) // Template ID
	buf = binary.BigEndian.AppendUint16(buf, 7)   // Field count

	// Field specifiers (4 bytes each: 2 for IE ID, 2 for length)
	// protocolIdentifier (IE 4) - 1 byte
	buf = binary.BigEndian.AppendUint16(buf, 4)
	buf = binary.BigEndian.AppendUint16(buf, 1)

	// sourceTransportPort (IE 7) - 2 bytes
	buf = binary.BigEndian.AppendUint16(buf, 7)
	buf = binary.BigEndian.AppendUint16(buf, 2)

	// sourceIPv4Address (IE 8) - 4 bytes
	buf = binary.BigEndian.AppendUint16(buf, 8)
	buf = binary.BigEndian.AppendUint16(buf, 4)

	// destinationTransportPort (IE 11) - 2 bytes
	buf = binary.BigEndian.AppendUint16(buf, 11)
	buf = binary.BigEndian.AppendUint16(buf, 2)

	// destinationIPv4Address (IE 12) - 4 bytes
	buf = binary.BigEndian.AppendUint16(buf, 12)
	buf = binary.BigEndian.AppendUint16(buf, 4)

	// octetDeltaCount (IE 1) - 8 bytes
	buf = binary.BigEndian.AppendUint16(buf, 1)
	buf = binary.BigEndian.AppendUint16(buf, 8)

	// packetDeltaCount (IE 2) - 8 bytes
	buf = binary.BigEndian.AppendUint16(buf, 2)
	buf = binary.BigEndian.AppendUint16(buf, 8)

	// Data Set Header (4 bytes) + Data Record (29 bytes) = 33 bytes
	buf = binary.BigEndian.AppendUint16(buf, 256) // Set ID = Template ID
	buf = binary.BigEndian.AppendUint16(buf, 33)  // Set length

	// Data Record
	// protocolIdentifier (1 byte) - 6=TCP, 17=UDP
	protocols := []byte{6, 17}
	buf = append(buf, protocols[rand.Intn(len(protocols))])

	// sourceTransportPort (2 bytes)
	srcPort := uint16(rand.Intn(65535-1024) + 1024)
	buf = binary.BigEndian.AppendUint16(buf, srcPort)

	// sourceIPv4Address (4 bytes)
	srcIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	buf = append(buf, srcIP...)

	// destinationTransportPort (2 bytes)
	dstPort := SamplePorts[rand.Intn(len(SamplePorts))]
	buf = binary.BigEndian.AppendUint16(buf, dstPort)

	// destinationIPv4Address (4 bytes)
	dstIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	buf = append(buf, dstIP...)

	// octetDeltaCount (8 bytes) - random bytes between 64 and 65535
	bytes := uint64(rand.Intn(65535-64) + 64)
	buf = binary.BigEndian.AppendUint64(buf, bytes)

	// packetDeltaCount (8 bytes) - random packets between 1 and 1000
	packets := uint64(rand.Intn(1000) + 1)
	buf = binary.BigEndian.AppendUint64(buf, packets)

	return buf
}
