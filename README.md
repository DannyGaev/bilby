# bilby
C2 Communication Via Reverse DNS Tunneling

bilby exploits the usual necessity for DNS abused by DNS tunneling, but encodes data into the IP addresses themselves. Returned hostnames from the bilby server are used to communicate commands for the client to execute. bilby currently supports basic C2 commands:
* ls command
* targeted data exfiltration

bilby is able to send three bytes at a time while communicating with the server, and as such is not suited for exfiltrating large files quickly; smaller files, and the packaged C2 commands are the best fit for this tool's usage.

Once it is running, the bilby client (**C**) will continuously send heartbeat data to the bilby server (*S*). *S* may optionally reply with commands for **C** to execute, but is not required to do so. Commands are specified in the following format:

    cmd.[action]-[type].[filename-or-filepath].[extension]

For example:

    cmd-exfil-single.test.txt

will tell **C** that it has received a command (cmd) to exfiltrate (exfil) the single file (single) test.txt (test.txt). 

If a path must be specified, the command format can be changed:

    cmd-exfil-path.home-dscully-test.txt

this tells **C** to use the path /home/dscully/test.txt for locating the file to be exfiltrated.

This command format, though, exposes the server's intentions very quickly. For that purpose, **C** can be modified prior to deployment to include explicit mappings between hardcoded command strings and custom-tailored strings that will be sent by the server. For instance,

    cmd-exfil-path.home-dscully-test.txt

becomes

    nginxplus-al-in.home-dscully-test.txt

Further specification can mask the targeted file(s):

    nginxplus-al-in.1f96-lb-github.com

though this requires that you know the name and extension of the targeted file in advance. Client hardcodings cannot be changed without rebuilding the application client-side.

On **C**'s side, each section of the received 'command' hostname is broken down as follows:

[0]     cmd

[1]		action-type

[2]		filename-or-filepath

[3]		extension

as such, each index is expected to hold certain values. This format must be followed for **C** to function correctly.