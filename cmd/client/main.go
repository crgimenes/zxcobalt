package main

import (
	"io"
	"log"
	"net"
	"os"
)

const socketPath = "/tmp/remoteshell.sock"

func main() {
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to connect to %s: %v", socketPath, err)
	}
	defer conn.Close()

	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil {
			log.Printf("Error sending input: %v", err)
		}
	}()
	_, err = io.Copy(os.Stdout, conn)
	if err != nil {
		log.Printf("Error receiving output: %v", err)
	}
}
