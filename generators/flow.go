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
	buf = binary.BigEndian.AppendUint32(buf, 1)           // Sample type: flow sample
	buf = binary.BigEndian.AppendUint32(buf, 88)          // Sample length

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
	buf = binary.BigEndian.AppendUint32(buf, 1)    // Record type: raw packet header
	buf = binary.BigEndian.AppendUint32(buf, 48)   // Record length

	// Raw packet header format
	buf = binary.BigEndian.AppendUint32(buf, 1)    // Header protocol (1 = Ethernet)
	buf = binary.BigEndian.AppendUint32(buf, 1500) // Frame length
	buf = binary.BigEndian.AppendUint32(buf, 0)    // Stripped bytes
	buf = binary.BigEndian.AppendUint32(buf, 34)   // Header length

	// Minimal Ethernet + IP header (34 bytes)
	// Ethernet header (14 bytes)
	buf = append(buf, 0x00, 0x11, 0x22, 0x33, 0x44, 0x55) // Dst MAC
	buf = append(buf, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb) // Src MAC
	buf = append(buf, 0x08, 0x00)                          // EtherType: IPv4

	// IP header (20 bytes)
	buf = append(buf, 0x45, 0x00)                          // Version/IHL, DSCP
	buf = binary.BigEndian.AppendUint16(buf, 1480)         // Total length
	buf = binary.BigEndian.AppendUint16(buf, uint16(rand.Uint32()&0xFFFF)) // ID
	buf = append(buf, 0x00, 0x00)                          // Flags/Fragment
	buf = append(buf, 64, 6)                               // TTL, Protocol (TCP)
	buf = append(buf, 0x00, 0x00)                          // Checksum (0 for fake)

	// Random source/dest IPs
	srcIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	dstIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	buf = append(buf, srcIP...)
	buf = append(buf, dstIP...)

	return buf
}

// GenerateIPFIXPacket generates a minimal valid IPFIX packet
// Structure based on RFC 7011
func GenerateIPFIXPacket() []byte {
	buf := make([]byte, 0, 128)

	// IPFIX Message Header (16 bytes)
	buf = binary.BigEndian.AppendUint16(buf, 10)           // Version 10 (IPFIX)
	buf = binary.BigEndian.AppendUint16(buf, 52)           // Length
	buf = binary.BigEndian.AppendUint32(buf, uint32(time.Now().Unix())) // Export time
	buf = binary.BigEndian.AppendUint32(buf, rand.Uint32()) // Sequence number
	buf = binary.BigEndian.AppendUint32(buf, 12345)         // Observation domain ID

	// Template Set Header (4 bytes)
	buf = binary.BigEndian.AppendUint16(buf, 2)    // Set ID = 2 (Template Set)
	buf = binary.BigEndian.AppendUint16(buf, 20)   // Set length

	// Template Record
	buf = binary.BigEndian.AppendUint16(buf, 256)  // Template ID
	buf = binary.BigEndian.AppendUint16(buf, 2)    // Field count

	// Field specifiers
	// sourceIPv4Address (IE 8)
	buf = binary.BigEndian.AppendUint16(buf, 8)    // IE ID
	buf = binary.BigEndian.AppendUint16(buf, 4)    // Length

	// destinationIPv4Address (IE 12)
	buf = binary.BigEndian.AppendUint16(buf, 12)   // IE ID
	buf = binary.BigEndian.AppendUint16(buf, 4)    // Length

	// Data Set Header (4 bytes)
	buf = binary.BigEndian.AppendUint16(buf, 256)  // Set ID = Template ID
	buf = binary.BigEndian.AppendUint16(buf, 12)   // Set length

	// Data Record (8 bytes - 2 IPv4 addresses)
	srcIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	dstIP := net.ParseIP(SampleIPs[rand.Intn(len(SampleIPs))]).To4()
	buf = append(buf, srcIP...)
	buf = append(buf, dstIP...)

	return buf
}
