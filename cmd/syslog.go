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

var syslogHost string
var syslogPort int
var syslogRate int
var syslogCount int
var syslogRFC string
var syslogType string

var syslogCmd = &cobra.Command{
	Use:   "syslog",
	Short: "Send fake syslog messages over UDP",
	Long: `Send fake syslog messages over UDP to test syslog ingestion.

Supports both RFC3164 and RFC5424 formats, with multiple message types:
  generic   - Standard auth/system messages (default)
  tms       - DDoS mitigation system logs (blocked_host events)
  firewall  - UFW/iptables style firewall logs
  ids       - Snort/Suricata IDS alert format

Example:
  fakedata syslog --host 127.0.0.1 --port 514 --rfc 3164
  fakedata syslog --host 127.0.0.1 --port 514 --rfc 5424
  fakedata syslog --host 127.0.0.1 --port 514 --type tms --rate 100
  fakedata syslog --host 127.0.0.1 --port 514 --type firewall
  fakedata syslog --host 127.0.0.1 --port 514 --type ids
`,
	RunE: runSyslog,
}

func init() {
	syslogCmd.Flags().StringVar(&syslogHost, "host", "127.0.0.1", "Target host")
	syslogCmd.Flags().IntVar(&syslogPort, "port", 514, "Target port")
	syslogCmd.Flags().IntVar(&syslogRate, "rate", 10, "Messages per second")
	syslogCmd.Flags().IntVar(&syslogCount, "count", 0, "Total messages to send (0 = unlimited)")
	syslogCmd.Flags().StringVar(&syslogRFC, "rfc", "3164", "Syslog RFC format (3164 or 5424)")
	syslogCmd.Flags().StringVar(&syslogType, "type", "generic", "Message type: generic, tms, firewall, ids")
}

func runSyslog(cmd *cobra.Command, args []string) error {
	if syslogRFC != "3164" && syslogRFC != "5424" {
		return fmt.Errorf("invalid RFC format: %s (must be 3164 or 5424)", syslogRFC)
	}

	// Validate message type
	validTypes := map[string]bool{"generic": true, "tms": true, "firewall": true, "ids": true}
	if !validTypes[syslogType] {
		return fmt.Errorf("invalid message type: %s (must be generic, tms, firewall, or ids)", syslogType)
	}

	addr := fmt.Sprintf("%s:%d", syslogHost, syslogPort)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", addr, err)
	}
	defer conn.Close()

	fmt.Printf("Sending fake syslog (type=%s, RFC%s) to UDP %s at %d msg/s\n", syslogType, syslogRFC, addr, syslogRate)
	if syslogCount > 0 {
		fmt.Printf("Will send %d messages total\n", syslogCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(syslogRate)
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
			var msg string
			switch syslogType {
			case "tms":
				msg = generators.GenerateTMSSyslog()
			case "firewall":
				msg = generators.GenerateFirewallSyslog()
			case "ids":
				msg = generators.GenerateIDSSyslog()
			default:
				msg = generators.GenerateSyslogMessage(syslogRFC)
			}

			if _, err := conn.Write([]byte(msg)); err != nil {
				fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d messages...\n", sent)
			}

			if syslogCount > 0 && sent >= syslogCount {
				fmt.Printf("Completed. Sent %d messages in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
