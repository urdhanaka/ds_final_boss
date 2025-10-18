package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
)

func printErr(err error) {
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	conn, err := net.Dial("unix", "/var/lib/libvirt/console-logs/b.sock")
	printErr(err)
	if err != nil {
		return
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("Error reading from socket: %v", err)
			break
		}
		fmt.Printf("VM Log: %s", line)
	}
}
