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

var tcpHost string
var tcpPort int
var tcpRate int
var tcpCount int

var tcpCmd = &cobra.Command{
	Use:   "tcp",
	Short: "Send fake JSON data over TCP",
	Long: `Send fake JSON data over TCP to test TCP ingestion.

Each message is a JSON object with fields like:
  - timestamp, source_ip, dest_ip, source_port, dest_port
  - username, action, status, process
  - bytes_sent, bytes_recv, duration_ms, session_id

Example:
  fakedata tcp --host 127.0.0.1 --port 5001 --rate 100
`,
	RunE: runTCP,
}

func init() {
	tcpCmd.Flags().StringVar(&tcpHost, "host", "127.0.0.1", "Target host")
	tcpCmd.Flags().IntVar(&tcpPort, "port", 5001, "Target port")
	tcpCmd.Flags().IntVar(&tcpRate, "rate", 10, "Messages per second")
	tcpCmd.Flags().IntVar(&tcpCount, "count", 0, "Total messages to send (0 = unlimited)")
}

func runTCP(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", tcpHost, tcpPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	fmt.Printf("Sending fake JSON data to TCP %s at %d msg/s\n", addr, tcpRate)
	if tcpCount > 0 {
		fmt.Printf("Will send %d messages total\n", tcpCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(tcpRate)
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
				return fmt.Errorf("connection closed: %w", err)
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d messages...\n", sent)
			}

			if tcpCount > 0 && sent >= tcpCount {
				fmt.Printf("Completed. Sent %d messages in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
