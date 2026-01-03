# ByteFreezer Fakedata - Release Notes

## 2025-12-29

### Features
- **New syslog message types**: Added specialized syslog generators for security testing:
  - `tms` - DDoS mitigation system logs (TMS blocked_host events)
  - `firewall` - UFW/iptables style firewall logs
  - `ids` - Snort/Suricata IDS alert format

### Usage
```bash
# TMS DDoS mitigation logs
fakedata syslog --type tms --host 127.0.0.1 --port 514 --rate 100

# Firewall logs
fakedata syslog --type firewall --host 127.0.0.1 --port 514 --rate 100

# IDS alerts
fakedata syslog --type ids --host 127.0.0.1 --port 514 --rate 100
```

### Sample Output

**TMS Messages:**
```
<14>Dec 29 12:44:27 tms6ash tms[24536]: blocked_host addr=186.116.106.21, src_port=40239, dst_port=23, protocol=6, mitigation=AC_ATHEER_GRE_2015-12-22_Always-On, prefixes=212.70.50.0/24, countermeasure=filter, reason=filter_list_0, rule=0, blacklisted=no
```

**Firewall Messages:**
```
<13>Dec 29 12:44:27 fw-18 kernel[3376]: [UFW BLOCK] IN=eth1 OUT=eth1 SRC=80.82.77.33 DST=10.0.0.50 LEN=1236 TOS=0x00 PREC=0x00 TTL=15 ID=4147 PROTO=TCP SPT=17332 DPT=1433
```

**IDS Messages:**
```
<10>Dec 29 12:44:27 ids-01 snort[5648]: [1:5186636:3] ET TROJAN Known Malware CnC {TCP} 91.240.118.173:59250 -> 172.16.0.11:11211
```

### Technical Changes
- Added `generators/syslog.go` with:
  - `GenerateTMSSyslog()` - TMS/DDoS mitigation format
  - `GenerateFirewallSyslog()` - UFW firewall format
  - `GenerateIDSSyslog()` - Snort IDS format
- Updated `cmd/syslog.go` with `--type` flag
- Added realistic malicious IP addresses and attack signatures

### Files Modified
- `generators/syslog.go` - New file with syslog generators
- `cmd/syslog.go` - Added --type flag for message type selection
