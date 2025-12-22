// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package cmd

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytefreezer/fakedata/generators"
	"github.com/spf13/cobra"
)

var ipfixHost string
var ipfixPort int
var ipfixRate int
var ipfixCount int

var ipfixCmd = &cobra.Command{
	Use:   "ipfix",
	Short: "Send fake IPFIX packets over UDP",
	Long: `Send fake IPFIX (RFC 7011) packets over UDP to test IPFIX ingestion.

Generates valid IPFIX packets with template and data sets containing
source/destination IP addresses.

Example:
  fakedata ipfix --host 127.0.0.1 --port 4739
`,
	RunE: runIPFIX,
}

func init() {
	ipfixCmd.Flags().StringVar(&ipfixHost, "host", "127.0.0.1", "Target host")
	ipfixCmd.Flags().IntVar(&ipfixPort, "port", 4739, "Target port")
	ipfixCmd.Flags().IntVar(&ipfixRate, "rate", 10, "Packets per second")
	ipfixCmd.Flags().IntVar(&ipfixCount, "count", 0, "Total packets to send (0 = unlimited)")
}

func runIPFIX(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", ipfixHost, ipfixPort)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	fmt.Printf("Sending fake IPFIX packets to UDP %s at %d pkt/s\n", addr, ipfixRate)
	if ipfixCount > 0 {
		fmt.Printf("Will send %d packets total\n", ipfixCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(ipfixRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sent := 0
	startTime := time.Now()

	for {
		select {
		case <-sigChan:
			fmt.Printf("\nStopped. Sent %d packets in %v\n", sent, time.Since(startTime))
			return nil
		case <-ticker.C:
			packet := generators.GenerateIPFIXPacket()
			if _, err := conn.Write(packet); err != nil {
				fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d packets...\n", sent)
			}

			if ipfixCount > 0 && sent >= ipfixCount {
				fmt.Printf("Completed. Sent %d packets in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
