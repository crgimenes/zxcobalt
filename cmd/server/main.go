// server.go
package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"sync"
)

const socketPath = "/tmp/remoteshell.sock"

var (
	currentClient net.Conn
	clientMu      sync.Mutex
)

func main() {
	if _, err := os.Stat(socketPath); err == nil {
		os.Remove(socketPath)
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", socketPath, err)
	}
	defer listener.Close()
	log.Printf("Server listening on %s", socketPath)

	cmd := exec.Command("/bin/bash", "-i")
	shellStdin, err := cmd.StdinPipe()
	if err != nil {
		log.Fatalf("Failed to get shell stdin pipe: %v", err)
	}
	shellStdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Failed to get shell stdout pipe: %v", err)
	}
	cmd.Stderr = cmd.Stdout

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start shell: %v", err)
	}
	log.Printf("Shell process started with PID %d", cmd.Process.Pid)

	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := shellStdout.Read(buf)
			if err != nil {
				if err == io.EOF {
					log.Println("Shell process ended")
				} else {
					log.Printf("Error reading shell output: %v", err)
				}
				return
			}
			if n > 0 {
				clientMu.Lock()
				if currentClient != nil {
					_, werr := currentClient.Write(buf[:n])
					if werr != nil {
						log.Printf("Error writing to client: %v", werr)
						currentClient.Close()
						currentClient = nil
					}
				}
				clientMu.Unlock()
			}
		}
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		log.Printf("Client connected")

		clientMu.Lock()
		if currentClient != nil {
			log.Printf("Disconnecting previous client")
			currentClient.Close()
		}
		currentClient = conn
		clientMu.Unlock()

		go func(c net.Conn) {
			_, err := io.Copy(shellStdin, c)
			if err != nil {
				log.Printf("Error copying from client to shell: %v", err)
			}
			c.Close()
			clientMu.Lock()
			if currentClient == c {
				currentClient = nil
			}
			clientMu.Unlock()
			log.Printf("Client disconnected")
		}(conn)
	}
}
