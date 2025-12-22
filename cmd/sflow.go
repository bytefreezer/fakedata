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

var sflowHost string
var sflowPort int
var sflowRate int
var sflowCount int

var sflowCmd = &cobra.Command{
	Use:   "sflow",
	Short: "Send fake sFlow packets over UDP",
	Long: `Send fake sFlow v5 packets over UDP to test sFlow ingestion.

Generates valid sFlow v5 packets with flow samples containing
Ethernet/IP headers with random source/destination IPs.

Example:
  fakedata sflow --host 127.0.0.1 --port 6343
`,
	RunE: runSFlow,
}

func init() {
	sflowCmd.Flags().StringVar(&sflowHost, "host", "127.0.0.1", "Target host")
	sflowCmd.Flags().IntVar(&sflowPort, "port", 6343, "Target port")
	sflowCmd.Flags().IntVar(&sflowRate, "rate", 10, "Packets per second")
	sflowCmd.Flags().IntVar(&sflowCount, "count", 0, "Total packets to send (0 = unlimited)")
}

func runSFlow(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", sflowHost, sflowPort)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	fmt.Printf("Sending fake sFlow v5 packets to UDP %s at %d pkt/s\n", addr, sflowRate)
	if sflowCount > 0 {
		fmt.Printf("Will send %d packets total\n", sflowCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(sflowRate)
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
			packet := generators.GenerateSFlowPacket()
			if _, err := conn.Write(packet); err != nil {
				fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d packets...\n", sent)
			}

			if sflowCount > 0 && sent >= sflowCount {
				fmt.Printf("Completed. Sent %d packets in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
