// Licensed under Elastic License 2.0
// See LICENSE.txt for details

package generators

import (
	"fmt"
	"math/rand"
	"time"
)

// Real public IPs known for malicious activity (from threat intel feeds)
var MaliciousIPs = []string{
	// Known scanner/attacker IPs (anonymized but realistic ranges)
	"186.116.106.21", "186.116.106.45", "186.116.107.12",
	"45.227.255.206", "45.227.255.99", "45.227.254.11",
	"185.220.101.35", "185.220.101.42", "185.220.101.78",
	"89.248.167.131", "89.248.167.144", "89.248.165.89",
	"80.82.77.139", "80.82.77.33", "80.82.78.22",
	"71.6.135.131", "71.6.146.185", "71.6.146.130",
	"198.235.24.24", "198.235.24.156", "198.235.24.89",
	"62.102.148.68", "62.102.148.69", "62.102.148.70",
	"185.142.236.34", "185.142.236.35", "185.142.236.36",
	"45.95.169.139", "45.95.169.140", "45.95.169.141",
	"103.251.167.20", "103.251.167.21", "103.251.167.22",
	"194.26.29.120", "194.26.29.121", "194.26.29.122",
	"141.98.10.60", "141.98.10.61", "141.98.10.62",
	"179.43.175.20", "179.43.175.21", "179.43.175.22",
	"91.240.118.172", "91.240.118.173", "91.240.118.174",
}

// Protected network prefixes
var ProtectedPrefixes = []string{
	"212.70.50.0/24", "212.70.51.0/24", "212.70.52.0/24",
	"185.45.12.0/22", "185.45.16.0/22",
	"103.28.52.0/24", "103.28.53.0/24",
	"45.60.0.0/16", "45.61.0.0/16",
	"192.0.2.0/24", "198.51.100.0/24", "203.0.113.0/24",
	"64.233.160.0/19", "74.125.0.0/16",
	"157.240.0.0/16", "31.13.24.0/21",
}

// Mitigation names (realistic DDoS mitigation rule names)
var Mitigations = []string{
	"AC_ATHEER_GRE_2015-12-22_Always-On",
	"AC_GLOBAL_UDP_FLOOD_Protection",
	"AC_SYN_FLOOD_Mitigation_2023",
	"AC_DNS_AMP_Block_v2",
	"AC_NTP_AMP_Filter",
	"AC_MEMCACHED_Block",
	"AC_CHARGEN_Filter",
	"AC_SSDP_AMP_Block",
	"AC_CLDAP_Mitigation",
	"AC_RDP_BRUTE_Block",
	"AC_SSH_SCANNER_Filter",
	"AC_TELNET_BLOCK_2024",
	"AC_ICMP_FLOOD_Limit",
	"AC_HTTP_FLOOD_L7_v3",
	"AC_SLOWLORIS_Protection",
}

// Countermeasures
var Countermeasures = []string{
	"filter", "rate_limit", "blackhole", "scrub", "challenge", "drop", "tarpit",
}

// Filter reasons
var FilterReasons = []string{
	"filter_list_0", "filter_list_1", "filter_list_2",
	"geo_block", "reputation_block", "rate_exceeded",
	"signature_match", "anomaly_detected", "threshold_exceeded",
	"bogon_source", "spoofed_source", "amplification_attack",
}

// TMS hostnames
var TMSHostnames = []string{
	"tms6ash", "tms7bos", "tms8chi", "tms9dal", "tms10den",
	"tms11lax", "tms12mia", "tms13nyc", "tms14phx", "tms15sea",
	"tms16sjc", "tms17was", "tms18ams", "tms19fra", "tms20lon",
}

// Attack destination ports
var AttackPorts = []int{
	22, 23, 25, 53, 80, 123, 161, 443, 445, 1433, 1900, 3306, 3389, 5060, 5900, 6379, 8080, 11211, 27017,
}

// Protocols: 1=ICMP, 6=TCP, 17=UDP
var Protocols = []int{1, 6, 17}

// GenerateTMSSyslog generates a TMS (Threat Mitigation System) syslog message
// Format: <14>May 11 19:43:09 tms6ash tms[24536]: blocked_host addr=IP, src_port=N, dst_port=N, protocol=N, mitigation=NAME, prefixes=PREFIX, countermeasure=TYPE, reason=REASON, rule=N, blacklisted=BOOL
func GenerateTMSSyslog() string {
	now := time.Now()

	// Priority 14 = facility 1 (user) * 8 + severity 6 (info)
	priority := 14

	hostname := TMSHostnames[rand.Intn(len(TMSHostnames))]
	pid := rand.Intn(50000) + 10000

	srcIP := MaliciousIPs[rand.Intn(len(MaliciousIPs))]
	srcPort := rand.Intn(65535-1024) + 1024
	dstPort := AttackPorts[rand.Intn(len(AttackPorts))]
	protocol := Protocols[rand.Intn(len(Protocols))]
	mitigation := Mitigations[rand.Intn(len(Mitigations))]
	prefix := ProtectedPrefixes[rand.Intn(len(ProtectedPrefixes))]
	countermeasure := Countermeasures[rand.Intn(len(Countermeasures))]
	reason := FilterReasons[rand.Intn(len(FilterReasons))]
	rule := rand.Intn(10)
	blacklisted := "no"
	if rand.Float32() < 0.3 {
		blacklisted = "yes"
	}

	// RFC3164 format
	return fmt.Sprintf("<%d>%s %s tms[%d]: blocked_host addr=%s, src_port=%d, dst_port=%d, protocol=%d, mitigation=%s, prefixes=%s, countermeasure=%s, reason=%s, rule=%d, blacklisted=%s",
		priority,
		now.Format("Jan  2 15:04:05"),
		hostname,
		pid,
		srcIP,
		srcPort,
		dstPort,
		protocol,
		mitigation,
		prefix,
		countermeasure,
		reason,
		rule,
		blacklisted,
	)
}

// GenerateFirewallSyslog generates firewall-style syslog messages
func GenerateFirewallSyslog() string {
	now := time.Now()
	priority := []int{12, 13, 14}[rand.Intn(3)] // Various info/notice priorities

	hostname := fmt.Sprintf("fw-%02d", rand.Intn(20)+1)
	pid := rand.Intn(10000) + 1000

	srcIP := MaliciousIPs[rand.Intn(len(MaliciousIPs))]
	dstIP := SampleIPs[rand.Intn(len(SampleIPs))]
	srcPort := rand.Intn(65535-1024) + 1024
	dstPort := AttackPorts[rand.Intn(len(AttackPorts))]

	actions := []string{"DENY", "DROP", "REJECT", "BLOCK"}
	action := actions[rand.Intn(len(actions))]

	interfaces := []string{"eth0", "eth1", "wan0", "lan0", "dmz0"}
	inIface := interfaces[rand.Intn(len(interfaces))]
	outIface := interfaces[rand.Intn(len(interfaces))]

	protocols := []string{"TCP", "UDP", "ICMP"}
	proto := protocols[rand.Intn(len(protocols))]

	return fmt.Sprintf("<%d>%s %s kernel[%d]: [UFW %s] IN=%s OUT=%s SRC=%s DST=%s LEN=%d TOS=0x00 PREC=0x00 TTL=%d ID=%d PROTO=%s SPT=%d DPT=%d",
		priority,
		now.Format("Jan  2 15:04:05"),
		hostname,
		pid,
		action,
		inIface,
		outIface,
		srcIP,
		dstIP,
		rand.Intn(1500)+40,
		rand.Intn(64)+1,
		rand.Intn(65535),
		proto,
		srcPort,
		dstPort,
	)
}

// GenerateIDSSyslog generates IDS/IPS style alerts
func GenerateIDSSyslog() string {
	now := time.Now()
	priority := 10 // security/auth warning

	hostname := fmt.Sprintf("ids-%02d", rand.Intn(10)+1)

	srcIP := MaliciousIPs[rand.Intn(len(MaliciousIPs))]
	dstIP := SampleIPs[rand.Intn(len(SampleIPs))]
	srcPort := rand.Intn(65535-1024) + 1024
	dstPort := AttackPorts[rand.Intn(len(AttackPorts))]

	signatures := []string{
		"ET SCAN Potential SSH Scan",
		"ET EXPLOIT Possible SQL Injection Attempt",
		"ET TROJAN Known Malware CnC",
		"ET DOS Possible NTP DDoS Amplification",
		"ET POLICY Outbound SSH Connection",
		"ET SCAN Nmap OS Detection Probe",
		"ET ATTACK_RESPONSE Suspicious 200 OK",
		"ET WEB_SERVER SQL Injection Attempt",
		"ET MALWARE Ransomware CnC Beacon",
		"ET SCAN Masscan Detected",
	}
	sig := signatures[rand.Intn(len(signatures))]

	sid := rand.Intn(9000000) + 1000000
	rev := rand.Intn(10) + 1

	return fmt.Sprintf("<%d>%s %s snort[%d]: [1:%d:%d] %s {TCP} %s:%d -> %s:%d",
		priority,
		now.Format("Jan  2 15:04:05"),
		hostname,
		rand.Intn(10000)+1000,
		sid,
		rev,
		sig,
		srcIP,
		srcPort,
		dstIP,
		dstPort,
	)
}
