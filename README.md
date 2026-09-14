# bilby
C2 Communication Via Reverse DNS Tunneling

bilby exploits the usual necessity for DNS abused by DNS tunneling, but encodes data into the IP addresses themselves. Returned hostnames from the bilby server are used to communicate commands for the client to execute. bilby currently supports basic C2 commands:
    - ls command
    - targeted data exfiltration

bilby is able to send three bytes at a time while communicating with the server, and as such is not suited for exfiltrating large files quickly; smaller files, and the packaged C2 commands are the best fit for this tool's usage.