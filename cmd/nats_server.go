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
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/spf13/cobra"
)

var natsServerPort int
var natsServerSubject string
var natsServerRate int
var natsServerCount int

var natsServerCmd = &cobra.Command{
	Use:   "nats-server",
	Short: "Run embedded NATS server and publish fake data (no external dependencies)",
	Long: `Start an embedded NATS server and publish fake JSON messages to it.

This is useful for testing ByteFreezer proxy NATS plugin without installing
any external NATS infrastructure. Configure your proxy to connect to this
embedded server.

Example:
  # Start embedded NATS on port 4222, publish to "events.test"
  fakedata nats-server --port 4222 --subject events.test --rate 100

  # Configure proxy to consume from nats://localhost:4222 subject "events.test"
`,
	RunE: runNATSServer,
}

func init() {
	natsServerCmd.Flags().IntVar(&natsServerPort, "port", 4222, "NATS server port")
	natsServerCmd.Flags().StringVar(&natsServerSubject, "subject", "bytefreezer.events", "Subject to publish to")
	natsServerCmd.Flags().IntVar(&natsServerRate, "rate", 10, "Messages per second")
	natsServerCmd.Flags().IntVar(&natsServerCount, "count", 0, "Total messages to send (0 = unlimited)")
}

func runNATSServer(cmd *cobra.Command, args []string) error {
	// Create embedded NATS server
	opts := &server.Options{
		Host:           "0.0.0.0",
		Port:           natsServerPort,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 4096,
	}

	ns, err := server.NewServer(opts)
	if err != nil {
		return fmt.Errorf("failed to create NATS server: %w", err)
	}

	// Start server in background
	go ns.Start()

	// Wait for server to be ready
	if !ns.ReadyForConnections(10 * time.Second) {
		return fmt.Errorf("NATS server failed to start within timeout")
	}

	fmt.Printf("Embedded NATS server started on port %d\n", natsServerPort)
	fmt.Printf("Configure proxy to connect to: nats://localhost:%d\n", natsServerPort)
	fmt.Printf("Publishing to subject: %s\n", natsServerSubject)

	// Connect to embedded server
	nc, err := nats.Connect(fmt.Sprintf("nats://localhost:%d", natsServerPort))
	if err != nil {
		ns.Shutdown()
		return fmt.Errorf("failed to connect to embedded NATS: %w", err)
	}
	defer nc.Close()

	fmt.Printf("Sending fake JSON at %d msg/s\n", natsServerRate)
	if natsServerCount > 0 {
		fmt.Printf("Will send %d messages total\n", natsServerCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(natsServerRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sent := 0
	startTime := time.Now()

	for {
		select {
		case <-sigChan:
			fmt.Printf("\nStopping...\n")
			nc.Flush()
			ns.Shutdown()
			fmt.Printf("Sent %d messages in %v\n", sent, time.Since(startTime))
			return nil
		case <-ticker.C:
			event, err := generators.GenerateJSONEvent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating event: %v\n", err)
				continue
			}
			if err := nc.Publish(natsServerSubject, event); err != nil {
				fmt.Fprintf(os.Stderr, "Error publishing: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d messages...\n", sent)
			}

			if natsServerCount > 0 && sent >= natsServerCount {
				nc.Flush()
				// Keep server running briefly so proxy can consume
				fmt.Printf("Sent %d messages. Server will continue running for consumption.\n", sent)
				fmt.Println("Press Ctrl+C to stop server")
				// Wait for signal
				<-sigChan
				ns.Shutdown()
				fmt.Printf("Completed in %v\n", time.Since(startTime))
				return nil
			}
		}
	}
}
