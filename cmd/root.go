// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fakedata",
	Short: "Generate fake data for testing data ingestion pipelines",
	Long: `Fakedata - Synthetic Data Generator for Testing Data Pipelines

Generate synthetic security/network events and send them over any protocol
your pipeline supports. Single binary, no dependencies.

SUPPORTED PROTOCOLS
  udp         Send JSON events over UDP
  tcp         Send JSON events over TCP
  syslog      Send syslog messages (RFC 3164 or RFC 5424)
  sflow       Send sFlow v5 packets
  ipfix       Send IPFIX/NetFlow packets
  nats        Publish to external NATS server
  nats-server Run embedded NATS server + publish (no Docker needed!)
  kafka       Produce to Kafka/Redpanda
  sqs         Send to AWS SQS (supports LocalStack)
  kinesis     Put to AWS Kinesis (supports LocalStack)

COMMON FLAGS
  --host      Target host/IP address
  --port      Target port number
  --rate      Messages per second (default: 10)
  --count     Total messages to send, 0 = unlimited (default: 0)

EXAMPLES (copy-paste ready)

  Network protocols:
    fakedata udp --host 127.0.0.1 --port 5000 --rate 100
    fakedata tcp --host 127.0.0.1 --port 5001 --rate 100
    fakedata syslog --host 127.0.0.1 --port 514 --rfc 3164 --rate 100
    fakedata syslog --host 127.0.0.1 --port 514 --rfc 5424 --rate 100
    fakedata sflow --host 127.0.0.1 --port 6343 --rate 100
    fakedata ipfix --host 127.0.0.1 --port 4739 --rate 100

  Message queues:
    fakedata nats-server --port 4222 --subject events --rate 100
    fakedata nats --servers nats://localhost:4222 --subject events --rate 100
    fakedata kafka --brokers localhost:9092 --topic events --rate 100
    fakedata sqs --queue-url http://localhost:4566/000000000000/q --endpoint http://localhost:4566 --rate 100
    fakedata kinesis --stream test-stream --endpoint http://localhost:4566 --rate 100

  With count (send N messages then stop):
    fakedata udp --host 127.0.0.1 --port 5000 --rate 100 --count 1000

  Load testing:
    fakedata udp --host 127.0.0.1 --port 5000 --rate 10000 --count 600000
    fakedata udp --host 127.0.0.1 --port 5000 --rate 50000 --count 500000

SAMPLE JSON EVENT
  {
    "timestamp": "2024-01-15T10:23:45.123456789Z",
    "source_ip": "192.168.1.100",
    "dest_ip": "8.8.8.8",
    "source_port": 54321,
    "dest_port": 443,
    "username": "admin",
    "action": "login",
    "status": "success",
    "process": "sshd",
    "bytes_sent": 1234,
    "bytes_recv": 5678,
    "duration_ms": 150,
    "session_id": "sess_123456789"
  }

For more details on a specific command:
  fakedata <command> --help

Project: https://github.com/bytefreezer/fakedata
License: Elastic License 2.0
`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(udpCmd)
	rootCmd.AddCommand(tcpCmd)
	rootCmd.AddCommand(syslogCmd)
	rootCmd.AddCommand(sflowCmd)
	rootCmd.AddCommand(ipfixCmd)
	rootCmd.AddCommand(natsCmd)
	rootCmd.AddCommand(natsServerCmd)
	rootCmd.AddCommand(kafkaCmd)
	rootCmd.AddCommand(sqsCmd)
	rootCmd.AddCommand(kinesisCmd)
}
