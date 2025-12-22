// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package generators

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/bytedance/sonic"
)

// SampleIPs contains realistic IP addresses for testing
var SampleIPs = []string{
	"192.168.1.100", "192.168.1.101", "192.168.1.102",
	"10.0.0.50", "10.0.0.51", "10.0.0.52",
	"172.16.0.10", "172.16.0.11", "172.16.0.12",
	"8.8.8.8", "1.1.1.1", "208.67.222.222",
	"104.16.132.229", "151.101.1.140", "185.199.108.153",
}

// SampleUsernames contains sample usernames
var SampleUsernames = []string{
	"admin", "root", "user1", "jsmith", "alice", "bob",
	"sysadmin", "guest", "operator", "developer",
}

// SampleActions contains sample action types
var SampleActions = []string{
	"login", "logout", "create", "update", "delete",
	"read", "write", "execute", "connect", "disconnect",
}

// SampleStatuses contains sample status values
var SampleStatuses = []string{
	"success", "failed", "pending", "timeout", "error",
}

// SampleProcesses contains sample process names
var SampleProcesses = []string{
	"sshd", "nginx", "apache2", "docker", "kubelet",
	"systemd", "cron", "bash", "python3", "node",
}

// GenerateJSONEvent generates a random JSON event for testing
func GenerateJSONEvent() ([]byte, error) {
	event := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
		"source_ip":   SampleIPs[rand.Intn(len(SampleIPs))],
		"dest_ip":     SampleIPs[rand.Intn(len(SampleIPs))],
		"source_port": rand.Intn(65535-1024) + 1024,
		"dest_port":   []int{22, 80, 443, 3306, 5432, 8080, 8443}[rand.Intn(7)],
		"username":    SampleUsernames[rand.Intn(len(SampleUsernames))],
		"action":      SampleActions[rand.Intn(len(SampleActions))],
		"status":      SampleStatuses[rand.Intn(len(SampleStatuses))],
		"process":     SampleProcesses[rand.Intn(len(SampleProcesses))],
		"bytes_sent":  rand.Intn(100000),
		"bytes_recv":  rand.Intn(100000),
		"duration_ms": rand.Intn(5000),
		"session_id":  fmt.Sprintf("sess_%d", rand.Int63()),
	}

	return sonic.Marshal(event)
}

// GenerateSyslogMessage generates a syslog message
func GenerateSyslogMessage(rfc string) string {
	now := time.Now()
	hostname := fmt.Sprintf("server%d", rand.Intn(10)+1)
	process := SampleProcesses[rand.Intn(len(SampleProcesses))]
	pid := rand.Intn(65535)

	// Priority: facility * 8 + severity
	// facility: 1 (user-level), 4 (security/auth)
	// severity: 4 (warning), 5 (notice), 6 (info)
	priority := ([]int{1, 4}[rand.Intn(2)] * 8) + rand.Intn(3) + 4

	messages := []string{
		fmt.Sprintf("User %s logged in from %s", SampleUsernames[rand.Intn(len(SampleUsernames))], SampleIPs[rand.Intn(len(SampleIPs))]),
		fmt.Sprintf("Connection from %s port %d", SampleIPs[rand.Intn(len(SampleIPs))], rand.Intn(65535)),
		fmt.Sprintf("Failed password for %s from %s", SampleUsernames[rand.Intn(len(SampleUsernames))], SampleIPs[rand.Intn(len(SampleIPs))]),
		fmt.Sprintf("Session opened for user %s", SampleUsernames[rand.Intn(len(SampleUsernames))]),
		fmt.Sprintf("Session closed for user %s", SampleUsernames[rand.Intn(len(SampleUsernames))]),
		fmt.Sprintf("Accepted publickey for %s from %s", SampleUsernames[rand.Intn(len(SampleUsernames))], SampleIPs[rand.Intn(len(SampleIPs))]),
		fmt.Sprintf("Process %s started with PID %d", process, pid),
		fmt.Sprintf("Service %s reloaded", process),
	}
	msg := messages[rand.Intn(len(messages))]

	switch rfc {
	case "5424":
		// RFC5424: <PRI>VERSION TIMESTAMP HOSTNAME APP-NAME PROCID MSGID [STRUCTURED-DATA] MSG
		return fmt.Sprintf("<%d>1 %s %s %s %d - - %s",
			priority,
			now.Format("2006-01-02T15:04:05.000000Z07:00"),
			hostname,
			process,
			pid,
			msg,
		)
	default:
		// RFC3164: <PRI>TIMESTAMP HOSTNAME TAG: MESSAGE
		return fmt.Sprintf("<%d>%s %s %s[%d]: %s",
			priority,
			now.Format("Jan  2 15:04:05"),
			hostname,
			process,
			pid,
			msg,
		)
	}
}
