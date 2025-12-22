// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bytefreezer/fakedata/generators"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

var natsServers string
var natsSubject string
var natsRate int
var natsCount int

var natsCmd = &cobra.Command{
	Use:   "nats",
	Short: "Publish fake JSON messages to NATS",
	Long: `Publish fake JSON messages to a NATS server.

NATS is very lightweight and can run locally for testing:
  docker run -p 4222:4222 nats:latest

Example:
  fakedata nats --servers nats://localhost:4222 --subject events.test --rate 100
`,
	RunE: runNATS,
}

func init() {
	natsCmd.Flags().StringVar(&natsServers, "servers", "nats://localhost:4222", "NATS server URL(s), comma-separated")
	natsCmd.Flags().StringVar(&natsSubject, "subject", "bytefreezer.events", "Subject to publish to")
	natsCmd.Flags().IntVar(&natsRate, "rate", 10, "Messages per second")
	natsCmd.Flags().IntVar(&natsCount, "count", 0, "Total messages to send (0 = unlimited)")
}

func runNATS(cmd *cobra.Command, args []string) error {
	// Connect to NATS
	nc, err := nats.Connect(natsServers,
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2*time.Second),
		nats.DisconnectErrHandler(func(_ *nats.Conn, err error) {
			if err != nil {
				fmt.Fprintf(os.Stderr, "NATS disconnected: %v\n", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			fmt.Printf("NATS reconnected to %s\n", nc.ConnectedUrl())
		}),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS at %s: %w", natsServers, err)
	}
	defer nc.Close()

	fmt.Printf("Connected to NATS at %s\n", nc.ConnectedUrl())
	fmt.Printf("Publishing fake JSON to subject '%s' at %d msg/s\n", natsSubject, natsRate)
	if natsCount > 0 {
		fmt.Printf("Will send %d messages total\n", natsCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(natsRate)
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
			event, err := generators.GenerateJSONEvent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating event: %v\n", err)
				continue
			}
			if err := nc.Publish(natsSubject, event); err != nil {
				fmt.Fprintf(os.Stderr, "Error publishing: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d messages...\n", sent)
			}

			if natsCount > 0 && sent >= natsCount {
				// Flush to ensure all messages are sent
				if err := nc.Flush(); err != nil {
					fmt.Fprintf(os.Stderr, "Error flushing: %v\n", err)
				}
				fmt.Printf("Completed. Sent %d messages in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
