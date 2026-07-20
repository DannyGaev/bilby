# bilby
Data Exfiltration Via Reverse DNS Tunneling

bibly exploits the usual necessity for DNS abused by DNS tunneling, but encodes data into the IP addresses themselves. There are currently two modes:
* Single-byte exfiltration
* Triple-byte exfiltration

Single-byte mode offers the lowest-throughput data exfiltration of the two, and as such requires far more time than triple-byte mode. This repository will be updated with octet-byte exfiltration via IPv6.