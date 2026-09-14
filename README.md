# bilby
Data Exfiltration Via Reverse DNS Tunneling

bilby exploits the usual necessity for DNS abused by DNS tunneling, but encodes data into the IP addresses themselves. Returned hostnames from the bilby server are used to communicate commands for the client to execute. bilby currently supports
* Data exfiltration
* Basic C2 Commands:
    - ls
    - Choosing what data to exfiltrate