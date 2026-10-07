package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func main() {
	in := bufio.NewReader(os.Stdin)
	replacement_mappings := map[string]string{"..": "80", ".": "79", "|": "78", ";": "77", "'": "76", ">": "75", "/": "74", "-": "72"}
	for true {
		fmt.Printf("[?] bash command [b] or exfiltration command [e]: ")
		command_type, _ := in.ReadString('\n')
		cleaned_command := strings.TrimRight(command_type, "\r\n")
		switch string(cleaned_command) {
		case "b":
			fmt.Printf("[?] Enter the command: ")
			bash_command, _ := in.ReadString('\n')
			sections := strings.Split(bash_command, " ")
			primary_command := sections[0]
			args_and_cont := strings.Join(sections[1:], " ")

			for k, v := range replacement_mappings {
				if strings.Contains(args_and_cont, k) {
					args_and_cont = strings.Replace(args_and_cont, k, v, -1)
				}
			}
			args_and_cont = strings.Replace(args_and_cont, " ", "-", -1)
			hbt_command := fmt.Sprintf("cmd-bash-%v.%v", primary_command, args_and_cont)
			fmt.Printf("[!] Writing formatted command to the file 'hbt_command': %v\n", hbt_command)

			hbtc := []byte(hbt_command)
			f, _ := os.OpenFile("hbt_command", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if _, err := f.Write(hbtc); err != nil {
				log.Fatal(err)
				f.Close()
			}
			if err := f.Close(); err != nil {
				log.Fatal(err)
			}
		}
	}
}
