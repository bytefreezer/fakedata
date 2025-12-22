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

var udpHost string
var udpPort int
var udpRate int
var udpCount int

var udpCmd = &cobra.Command{
	Use:   "udp",
	Short: "Send fake JSON data over UDP",
	Long: `Send fake JSON data over UDP to test UDP ingestion.

Each message is a JSON object with fields like:
  - timestamp, source_ip, dest_ip, source_port, dest_port
  - username, action, status, process
  - bytes_sent, bytes_recv, duration_ms, session_id

Example:
  fakedata udp --host 127.0.0.1 --port 5000 --rate 100
`,
	RunE: runUDP,
}

func init() {
	udpCmd.Flags().StringVar(&udpHost, "host", "127.0.0.1", "Target host")
	udpCmd.Flags().IntVar(&udpPort, "port", 5000, "Target port")
	udpCmd.Flags().IntVar(&udpRate, "rate", 10, "Messages per second")
	udpCmd.Flags().IntVar(&udpCount, "count", 0, "Total messages to send (0 = unlimited)")
}

func runUDP(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", udpHost, udpPort)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	fmt.Printf("Sending fake JSON data to UDP %s at %d msg/s\n", addr, udpRate)
	if udpCount > 0 {
		fmt.Printf("Will send %d messages total\n", udpCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(udpRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sent := 0
	startTime := time.Now()

	for {
		select {
		case <-sigChan:
			fmt.Printf("\nStopped. Sent %d messages in %v\n", sent, time.Since(startTime))
			return nil
		case <-ticker.C:
			data, err := generators.GenerateJSONEvent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating event: %v\n", err)
				continue
			}

			data = append(data, '\n')
			if _, err := conn.Write(data); err != nil {
				fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d messages...\n", sent)
			}

			if udpCount > 0 && sent >= udpCount {
				fmt.Printf("Completed. Sent %d messages in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
