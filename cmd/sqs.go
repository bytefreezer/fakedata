// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/bytefreezer/fakedata/generators"
	"github.com/spf13/cobra"
)

var sqsQueueURL string
var sqsRegion string
var sqsEndpoint string
var sqsRate int
var sqsCount int

var sqsCmd = &cobra.Command{
	Use:   "sqs",
	Short: "Send fake JSON messages to AWS SQS",
	Long: `Send fake JSON messages to an AWS SQS queue.

For local testing, use LocalStack:
  docker run -p 4566:4566 localstack/localstack
  aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name test-queue

Example:
  # LocalStack
  fakedata sqs --queue-url http://localhost:4566/000000000000/test-queue --endpoint http://localhost:4566 --region us-east-1

  # AWS
  fakedata sqs --queue-url https://sqs.us-east-1.amazonaws.com/123456789/my-queue --region us-east-1
`,
	RunE: runSQS,
}

func init() {
	sqsCmd.Flags().StringVar(&sqsQueueURL, "queue-url", "", "SQS queue URL (required)")
	sqsCmd.Flags().StringVar(&sqsRegion, "region", "us-east-1", "AWS region")
	sqsCmd.Flags().StringVar(&sqsEndpoint, "endpoint", "", "Custom endpoint URL (for LocalStack)")
	sqsCmd.Flags().IntVar(&sqsRate, "rate", 10, "Messages per second")
	sqsCmd.Flags().IntVar(&sqsCount, "count", 0, "Total messages to send (0 = unlimited)")
	sqsCmd.MarkFlagRequired("queue-url")
}

func runSQS(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load AWS config
	var cfg aws.Config
	var err error

	if sqsEndpoint != "" {
		// Custom endpoint (LocalStack)
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(sqsRegion),
			config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "test",
					SecretAccessKey: "test",
				}, nil
			})),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(sqsRegion))
	}
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create SQS client
	var client *sqs.Client
	if sqsEndpoint != "" {
		client = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.BaseEndpoint = aws.String(sqsEndpoint)
		})
		fmt.Printf("Using custom endpoint: %s\n", sqsEndpoint)
	} else {
		client = sqs.NewFromConfig(cfg)
	}

	fmt.Printf("Sending fake JSON to SQS queue at %d msg/s\n", sqsRate)
	fmt.Printf("Queue URL: %s\n", sqsQueueURL)
	if sqsCount > 0 {
		fmt.Printf("Will send %d messages total\n", sqsCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(sqsRate)
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

			_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
				QueueUrl:    aws.String(sqsQueueURL),
				MessageBody: aws.String(string(event)),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error sending: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d messages...\n", sent)
			}

			if sqsCount > 0 && sent >= sqsCount {
				fmt.Printf("Completed. Sent %d messages in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
