// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/bytefreezer/fakedata/generators"
	"github.com/spf13/cobra"
)

var kinesisStream string
var kinesisRegion string
var kinesisEndpoint string
var kinesisRate int
var kinesisCount int

var kinesisCmd = &cobra.Command{
	Use:   "kinesis",
	Short: "Put fake JSON records to AWS Kinesis",
	Long: `Put fake JSON records to an AWS Kinesis Data Stream.

For local testing, use LocalStack:
  docker run -p 4566:4566 localstack/localstack
  aws --endpoint-url=http://localhost:4566 kinesis create-stream --stream-name test-stream --shard-count 1

Example:
  # LocalStack
  fakedata kinesis --stream test-stream --endpoint http://localhost:4566 --region us-east-1

  # AWS
  fakedata kinesis --stream my-stream --region us-east-1
`,
	RunE: runKinesis,
}

func init() {
	kinesisCmd.Flags().StringVar(&kinesisStream, "stream", "", "Kinesis stream name (required)")
	kinesisCmd.Flags().StringVar(&kinesisRegion, "region", "us-east-1", "AWS region")
	kinesisCmd.Flags().StringVar(&kinesisEndpoint, "endpoint", "", "Custom endpoint URL (for LocalStack)")
	kinesisCmd.Flags().IntVar(&kinesisRate, "rate", 10, "Records per second")
	kinesisCmd.Flags().IntVar(&kinesisCount, "count", 0, "Total records to send (0 = unlimited)")
	kinesisCmd.MarkFlagRequired("stream")
}

func runKinesis(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Load AWS config
	var cfg aws.Config
	var err error

	if kinesisEndpoint != "" {
		// Custom endpoint (LocalStack)
		cfg, err = config.LoadDefaultConfig(ctx,
			config.WithRegion(kinesisRegion),
			config.WithCredentialsProvider(aws.CredentialsProviderFunc(func(ctx context.Context) (aws.Credentials, error) {
				return aws.Credentials{
					AccessKeyID:     "test",
					SecretAccessKey: "test",
				}, nil
			})),
		)
	} else {
		cfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(kinesisRegion))
	}
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create Kinesis client
	var client *kinesis.Client
	if kinesisEndpoint != "" {
		client = kinesis.NewFromConfig(cfg, func(o *kinesis.Options) {
			o.BaseEndpoint = aws.String(kinesisEndpoint)
		})
		fmt.Printf("Using custom endpoint: %s\n", kinesisEndpoint)
	} else {
		client = kinesis.NewFromConfig(cfg)
	}

	fmt.Printf("Putting fake JSON records to Kinesis stream '%s' at %d rec/s\n", kinesisStream, kinesisRate)
	if kinesisCount > 0 {
		fmt.Printf("Will send %d records total\n", kinesisCount)
	} else {
		fmt.Println("Press Ctrl+C to stop")
	}

	// Setup signal handler
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	interval := time.Second / time.Duration(kinesisRate)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	sent := 0
	startTime := time.Now()

	for {
		select {
		case <-sigChan:
			fmt.Printf("\nStopped. Sent %d records in %v\n", sent, time.Since(startTime))
			return nil
		case <-ticker.C:
			event, err := generators.GenerateJSONEvent()
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error generating event: %v\n", err)
				continue
			}

			// Use random partition key for distribution across shards
			partitionKey := fmt.Sprintf("pk-%d", rand.Intn(1000))

			_, err = client.PutRecord(ctx, &kinesis.PutRecordInput{
				StreamName:   aws.String(kinesisStream),
				Data:         event,
				PartitionKey: aws.String(partitionKey),
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error putting record: %v\n", err)
				continue
			}

			sent++
			if sent%1000 == 0 {
				fmt.Printf("Sent %d records...\n", sent)
			}

			if kinesisCount > 0 && sent >= kinesisCount {
				fmt.Printf("Completed. Sent %d records in %v\n", sent, time.Since(startTime))
				return nil
			}
		}
	}
}
