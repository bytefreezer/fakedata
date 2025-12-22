// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/bytefreezer/fakedata/generators"
	"github.com/spf13/cobra"
)

var kafkaBrokers string
var kafkaTopic string
var kafkaRate int
var kafkaCount int

var kafkaCmd = &cobra.Command{
	Use:   "kafka",
	Short: "Produce fake JSON messages to Kafka",
	Long: `Produce fake JSON messages to a Kafka topic.

For local testing, use Redpanda (Kafka-compatible, lighter weight):
  docker run -p 9092:9092 vectorized/redpanda

Example:
  fakedata kafka --brokers localhost:9092 --topic events --rate 100
`,
	RunE: runKafka,
}

func init() {
	kafkaCmd.Flags().StringVar(&kafkaBrokers, "brokers", "localhost:9092", "Kafka broker addresses, comma-separated")
	kafkaCmd.Flags().StringVar(&kafkaTopic, "topic", "bytefreezer-events", "Topic to produce to")
	kafkaCmd.Flags().IntVar(&kafkaRate, "rate", 10, "Messages per second")
	kafkaCmd.Flags().IntVar(&kafkaCount, "count", 0, "Total messages to send (0 = unlimited)")
}

func runKafka(cmd *cobra.Command, args []string) error {
	brokerList := strings.Split(kafkaBrokers, ",")

	// Create Kafka producer config
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Version = sarama.V2_8_0_0

	// Create sync producer
	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %w", err)
	}
	defer producer.Close()

	fmt.Printf("Connected to Kafka brokers: %v\n", brokerList)
	fmt.Printf("Producing fake JSON to topic '%s' at %d msg/s\n", kafkaTopic, kafkaRate)
	if kafkaCount > 0 {
		fmt.Printf("Will send %d messages total\n", kafkaCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(kafkaRate)
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

			msg := &sarama.ProducerMessage{
				Topic: kafkaTopic,
				Value: sarama.ByteEncoder(event),
			}

			_, _, err = producer.SendMessage(msg)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error producing: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d messages...\n", sent)
			}

			if kafkaCount > 0 && sent >= kafkaCount {
				fmt.Printf("Completed. Sent %d messages in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
