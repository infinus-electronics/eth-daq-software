package server

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	BUFFER_SIZE = 10 * 1024 * 1024 // 10MB
)

type DataBuffer struct {
	port          int
	clientIP      string
	buffer        []byte
	mu            sync.Mutex
	bytesReceived int64
	lastCheck     time.Time
	rate          float64
}

func NewDataBuffer(port int, clientIP string) *DataBuffer {
	return &DataBuffer{
		port:      port,
		clientIP:  SanitizeFilename(clientIP),
		buffer:    make([]byte, 0, BUFFER_SIZE),
		lastCheck: time.Now(),
	}
}

func (db *DataBuffer) AddData(data []byte) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.buffer = append(db.buffer, data...)
	db.bytesReceived += int64(len(data))

	elapsed := time.Since(db.lastCheck).Seconds()
	if elapsed >= 1.0 {
		rate := float64(db.bytesReceived) / elapsed / 1024 / 1024 // MB/s
		db.rate = rate
		fmt.Printf("Port %d - %s Rate: %.2f MB/s\n", db.port, db.clientIP, rate)
		db.bytesReceived = 0
		db.lastCheck = time.Now()
	}

	if len(db.buffer) >= BUFFER_SIZE {
		db.Flush()
	}
}

func (db *DataBuffer) Flush() {
	if len(db.buffer) == 0 {
		return
	}

	// Copy the buffer data while the mutex is held
	data := make([]byte, len(db.buffer))
	copy(data, db.buffer)

	// Generate filename and reset buffer immediately
	filename := fmt.Sprintf("port%d_%s_%d.bin",
		db.port,
		db.clientIP,
		time.Now().UnixNano(),
	)
	db.buffer = make([]byte, 0, BUFFER_SIZE)

	// Handle write asynchronously
	go func(data []byte, filename string) {
		err := os.WriteFile(filepath.Join("data", filename), data, 0644)
		if err != nil {
			fmt.Printf("Failed to write file: %v\n", err)
		} else {
			fmt.Printf("Written %d bytes to %s\n", len(data), filename)
		}
	}(data, filename)
}

func StartListener(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Printf("Failed to start server on port %d: %v\n", port, err)
		return
	}
	defer listener.Close()

	fmt.Printf("TCP Server listening on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection on port %d: %v\n", port, err)
			continue
		}

		clientIP := GetClientIP(conn.RemoteAddr())
		fmt.Printf("New connection on port %d from %s\n", port, clientIP)

		buffer := NewDataBuffer(port, clientIP)
		go HandleConnection(conn, buffer)
	}
}

func HandleConnection(conn net.Conn, buffer *DataBuffer) {
	defer func() {
		buffer.Flush()
		conn.Close()
		fmt.Printf("Connection closed from %s:%d\n", buffer.clientIP, buffer.port)
	}()

	chunk := make([]byte, 1048576) // 1MB chunks
	for {
		n, err := conn.Read(chunk)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Error reading from %s:%d: %v\n",
					buffer.clientIP,
					buffer.port,
					err,
				)
			}
			return
		}

		buffer.AddData(chunk[:n])
	}
}

func GetClientIP(addr net.Addr) string {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "unknown"
	}
	return host
}

func SanitizeFilename(ip string) string {
	// Replace characters that might be problematic in filenames
	return strings.NewReplacer(
		":", "_",
		".", "_",
		"[", "",
		"]", "",
	).Replace(ip)
}
