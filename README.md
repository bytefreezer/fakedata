# Fakedata - Synthetic Data Generator for Testing Data Pipelines

## The Problem

Testing data ingestion pipelines is painful:
- Setting up real data sources takes time
- Message queue infrastructure (Kafka, NATS, SQS) requires Docker or cloud resources
- Generating realistic test data at scale is tedious
- Validating multiple protocols means multiple tools

## The Solution

**Fakedata** is a single binary that generates synthetic security/network events and sends them over any protocol your pipeline supports.

## Supported Protocols

| Protocol | Use Case |
|----------|----------|
| UDP/TCP | Raw JSON ingestion |
| Syslog | RFC 3164 & RFC 5424 formats |
| sFlow | Network flow monitoring (v5) |
| IPFIX | NetFlow/IPFIX collectors |
| NATS | Message queue testing |
| Kafka | Stream processing |
| AWS SQS | Cloud queue ingestion |
| AWS Kinesis | Stream ingestion |

## Key Features

- **Zero Dependencies for NATS** - Embedded NATS server, no Docker needed
- **LocalStack Support** - Test SQS/Kinesis locally without AWS credentials
- **Configurable Rate & Count** - Control throughput from 1 to 50,000+ events/sec
- **Multiple Protocols** - Single tool for all your ingestion testing needs

## Use Cases

- **Pipeline Validation** - Verify your ingestion stack handles expected formats
- **Load Testing** - Stress test with configurable rates up to 50k+ events/sec
- **Development** - No need for production data during development
- **CI/CD** - Automated testing without external dependencies

---

## Installation

```bash
go build -o fakedata .
```

## Usage

### UDP (JSON data)
```bash
# Send 100 JSON events per second to UDP port 5000
fakedata udp --host 127.0.0.1 --port 5000 --rate 100

# Send exactly 1000 events then stop
fakedata udp --host 127.0.0.1 --port 5000 --rate 100 --count 1000
```

### TCP (JSON data)
```bash
# Send 100 JSON events per second to TCP port 5001
fakedata tcp --host 127.0.0.1 --port 5001 --rate 100
```

### Syslog
```bash
# Send RFC3164 syslog messages
fakedata syslog --host 127.0.0.1 --port 514 --rfc 3164

# Send RFC5424 syslog messages
fakedata syslog --host 127.0.0.1 --port 514 --rfc 5424
```

### sFlow
```bash
# Send sFlow v5 packets
fakedata sflow --host 127.0.0.1 --port 6343 --rate 100
```

### IPFIX
```bash
# Send IPFIX packets
fakedata ipfix --host 127.0.0.1 --port 4739 --rate 100
```

## Message Queue Generators

### NATS (Embedded Server - No Dependencies)

The simplest way to test NATS - runs an embedded NATS server, no Docker or external install needed:

```bash
# Start embedded NATS server on port 4222 and publish fake data
fakedata nats-server --port 4222 --subject bytefreezer.events --rate 100

# Configure your proxy to connect to nats://localhost:4222, subject "bytefreezer.events"
```

### NATS (External Server)
```bash
# If you have an existing NATS server
docker run -p 4222:4222 nats:latest

# Publish fake JSON to NATS
fakedata nats --servers nats://localhost:4222 --subject events.test --rate 100
```

### Kafka
```bash
# Start Redpanda locally (Kafka-compatible, lighter weight)
docker run -p 9092:9092 vectorized/redpanda

# Produce fake JSON to Kafka
fakedata kafka --brokers localhost:9092 --topic events --rate 100
```

### SQS (LocalStack)
```bash
# Start LocalStack
docker run -p 4566:4566 localstack/localstack

# Create queue
aws --endpoint-url=http://localhost:4566 sqs create-queue --queue-name test-queue

# Send fake JSON to SQS
fakedata sqs --queue-url http://localhost:4566/000000000000/test-queue --endpoint http://localhost:4566 --rate 100
```

### Kinesis (LocalStack)
```bash
# Start LocalStack
docker run -p 4566:4566 localstack/localstack

# Create stream
aws --endpoint-url=http://localhost:4566 kinesis create-stream --stream-name test-stream --shard-count 1

# Put fake JSON records to Kinesis
fakedata kinesis --stream test-stream --endpoint http://localhost:4566 --rate 100
```

## Load Testing

Test performance under load:

```bash
# Sustained load: 10,000 events/sec for 60 seconds
fakedata udp --host <target-ip> --port 5000 --rate 10000 --count 600000

# Burst test: 50,000 events/sec
fakedata udp --host <target-ip> --port 5000 --rate 50000 --count 500000

# Multi-protocol (run in separate terminals)
fakedata udp --host <target-ip> --port 5000 --rate 1000 &
fakedata tcp --host <target-ip> --port 5001 --rate 1000 &
fakedata syslog --host <target-ip> --port 514 --rate 1000 &
```

## Sample JSON Event

All generators produce JSON events with this structure:
```json
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
```

## Troubleshooting

### No Data Received
- Check target is listening: `ss -tuln | grep <port>`
- Verify firewall allows traffic: `iptables -L -n`
- Test network connectivity: `nc -vz <target-ip> <port>`

### High Drop Rate
- Reduce send rate with `--rate`
- Check system resources (CPU, memory, disk I/O)
- Increase target buffer sizes in configuration

## License

Licensed under Elastic License 2.0. See LICENSE.txt for details.

---

Built for the [ByteFreezer](https://bytefreezer.com) project - AI-native security data lake.
