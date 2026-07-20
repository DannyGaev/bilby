import re
import random
import socket
from dns import resolver, reversename, message, query, name, rrset, rdataclass, rdatatype, flags

server_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
server_socket.bind(('127.0.0.1', 49250))
recovered = ""

def create_response(data,address):
    global server_socket
    query = message.from_wire(data)
    response = message.make_response(query)
    response.flags |= flags.AA
    name_item = response.question[0].name
    cname_rdata = rrset.from_text(
        name_item,
        2,
        rdataclass.IN,
        rdatatype.PTR,
        'www.example.com.'
    )
    response.answer.append(cname_rdata) 
    response.id = query.id
    server_socket.sendto(response.to_wire(), address)

def find_payload(hex_message):
    hm_slice = hex_message[hex_message.find("10000010000000000000")+len("10000010000000000000")+1:]
    octets = []
    split_up = [hm_slice[i:i+2] for i in range(0, len(hm_slice), 2)]
    index = 0
    prior_index = 0
    adding = True
    while index < (len(split_up)):
        if split_up[index][0] == "0" and adding:
            octets.append("".join(split_up[prior_index:index]))
            if split_up[index][1] == "7":
                adding = False
            index+=1
            prior_index = index
        index+=1

    new_payload = []
    for y in range(0,len(octets)):
        full_decimal = ""
        for x in range(0,len(octets[y]),2):
            full_decimal+=octets[y][x+1]
        new_payload.insert(0,full_decimal)

    if new_payload[0] == "172":
        return True, new_payload[-1]
    else:
        return False, new_payload

old_stream_id = "0000"
while True:
    data, address = server_socket.recvfrom(512)
    data = data.upper()
    hex_message = data.hex()
    if "00000c0001" in hex_message:
        is_baseline, val = find_payload(hex_message)
        if is_baseline:
            new_baseline = val
            print(f"[!] NEW BASELINE: {new_baseline}")
        else:
            stream_id = hex_message[0:4]
            if stream_id != old_stream_id:
                if val[0] == "104":
                    print("[!] FULL DATA RECEIVED.")
                    create_response(data, address)
                    break
                else:
                    codes = val[0]
                    octet_index = 1
                    while octet_index < len(val):
                        if int(codes[octet_index-1:octet_index]) % 2 == 0:
                            recovered_byte = (int(val[octet_index]) + 255) - int(new_baseline)
                        else:
                            recovered_byte = int(val[octet_index]) - int(new_baseline)

                        hex_recovered_byte = str(hex(int(recovered_byte))).split("0x")[1]
                        if len(hex_recovered_byte)==1:
                            hex_recovered_byte = "0"+hex_recovered_byte

                        recovered += hex_recovered_byte
                        
                        with open("recovered","wb") as f:
                            f.write(bytes.fromhex(recovered))
                        
                        octet_index += 1
                    old_stream_id = stream_id 
        create_response(data, address)
