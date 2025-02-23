package server

import (
	"fmt"
	"io"
	"maps"
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

type Server struct {
	buffers     map[BufferKey]*DataBuffer
	buffersLock sync.RWMutex
	// Track IP addresses and their connection times
	connectedIPs     map[string]*IPConnection
	connectedIPsLock sync.RWMutex
}

// First, let's create a type for our composite key
type BufferKey struct {
	IP   string
	Port int
}

func NewServer() *Server {
	return &Server{
		buffers:      make(map[BufferKey]*DataBuffer),
		connectedIPs: make(map[string]*IPConnection),
	}
}

type IPConnection struct {
	FirstSeen   time.Time
	LastSeen    time.Time
	ActivePorts map[int]bool
	TotalBytes  int64
}

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

// GetRate returns the current transfer rate for this buffer
func (db *DataBuffer) GetRate() float64 {
	db.mu.Lock()
	defer db.mu.Unlock()
	return db.rate
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
	// db.mu.Lock()
	// defer db.mu.Unlock()
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

func (s *Server) StartListener(port int) {
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
		// Create composite key
		key := BufferKey{
			IP:   clientIP,
			Port: port,
		}

		// Store the buffer in the map using composite key
		s.buffersLock.Lock()
		s.buffers[key] = buffer
		s.buffersLock.Unlock()

		go s.HandleConnection(conn, buffer, key)
	}
}

// Modified HandleConnection to include the buffer key
func (s *Server) HandleConnection(conn net.Conn, buffer *DataBuffer, key BufferKey) {
	s.AddIPConnection(buffer.clientIP, buffer.port)

	defer func() {
		buffer.Flush()
		s.RemoveIPPort(buffer.clientIP, buffer.port)
		conn.Close()

		s.buffersLock.Lock()
		delete(s.buffers, key)
		s.buffersLock.Unlock()

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
		s.UpdateIPBytes(buffer.clientIP, int64(n))
	}
}

func GetClientIP(addr net.Addr) string {
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return "unknown"
	}
	return host
}

// Modified GetBufferRate to use the composite key
func (s *Server) GetBufferRate(ip string, port int) (float64, bool) {
	s.buffersLock.RLock()
	defer s.buffersLock.RUnlock()

	key := BufferKey{
		IP:   ip,
		Port: port,
	}

	if buffer, exists := s.buffers[key]; exists {
		return buffer.GetRate(), true
	}
	return 0, false
}

// Modified GetAllBufferRates to return rates with IP information
func (s *Server) GetAllBufferRates() map[string]float64 {
	s.buffersLock.RLock()
	defer s.buffersLock.RUnlock()

	rates := make(map[string]float64)
	for key, buffer := range s.buffers {
		rateKey := fmt.Sprintf("%s:%d", key.IP, key.Port)
		rates[rateKey] = buffer.GetRate()
	}
	return rates
}

// GetIPPortRate can now use the composite key directly
func (s *Server) GetIPPortRate(ip string, port int) (float64, bool) {
	s.buffersLock.RLock()
	defer s.buffersLock.RUnlock()

	key := BufferKey{
		IP:   SanitizeFilename(ip),
		Port: port,
	}

	if buffer, exists := s.buffers[key]; exists {
		return buffer.GetRate(), true
	}
	return 0, false
}

// GetAllIPPortRates remains largely the same but uses the new key structure
func (s *Server) GetAllIPPortRates() map[string]float64 {
	s.buffersLock.RLock()
	defer s.buffersLock.RUnlock()

	rates := make(map[string]float64)
	for key, buffer := range s.buffers {
		rateKey := fmt.Sprintf("%s:%d", key.IP, key.Port)
		rates[rateKey] = buffer.GetRate()
	}
	return rates
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

// AddIPConnection records or updates an IP connection
func (s *Server) AddIPConnection(ip string, port int) {
	s.connectedIPsLock.Lock()
	defer s.connectedIPsLock.Unlock()

	now := time.Now()
	sanitizedIP := SanitizeFilename(ip)

	if conn, exists := s.connectedIPs[sanitizedIP]; exists {
		conn.LastSeen = now
		conn.ActivePorts[port] = true
	} else {
		s.connectedIPs[sanitizedIP] = &IPConnection{
			FirstSeen:   now,
			LastSeen:    now,
			ActivePorts: map[int]bool{port: true},
			TotalBytes:  0,
		}
	}
}

// RemoveIPPort removes a port from an IP's active connections
func (s *Server) RemoveIPPort(ip string, port int) {
	s.connectedIPsLock.Lock()
	defer s.connectedIPsLock.Unlock()

	sanitizedIP := SanitizeFilename(ip)
	if conn, exists := s.connectedIPs[sanitizedIP]; exists {
		delete(conn.ActivePorts, port)

		// If no more active ports, remove the IP entirely
		if len(conn.ActivePorts) == 0 {
			delete(s.connectedIPs, sanitizedIP)
		}
	}
}

// UpdateIPBytes updates the total bytes transferred for an IP
func (s *Server) UpdateIPBytes(ip string, bytes int64) {
	s.connectedIPsLock.Lock()
	defer s.connectedIPsLock.Unlock()

	sanitizedIP := SanitizeFilename(ip)
	if conn, exists := s.connectedIPs[sanitizedIP]; exists {
		conn.TotalBytes += bytes
	}
}

// GetIPInfo returns information about a specific IP
func (s *Server) GetIPInfo(ip string) (*IPConnection, bool) {
	s.connectedIPsLock.RLock()
	defer s.connectedIPsLock.RUnlock()

	sanitizedIP := SanitizeFilename(ip)
	conn, exists := s.connectedIPs[sanitizedIP]
	if !exists {
		return nil, false
	}

	// Return a copy to prevent concurrent access issues
	return &IPConnection{
		FirstSeen:   conn.FirstSeen,
		LastSeen:    conn.LastSeen,
		ActivePorts: maps.Clone(conn.ActivePorts),
		TotalBytes:  conn.TotalBytes,
	}, true
}

// GetAllConnectedIPs returns information about all connected IPs
func (s *Server) GetAllConnectedIPs() map[string]*IPConnection {
	s.connectedIPsLock.RLock()
	defer s.connectedIPsLock.RUnlock()

	// Create a deep copy of the map
	result := make(map[string]*IPConnection)
	for ip, conn := range s.connectedIPs {
		result[ip] = &IPConnection{
			FirstSeen:   conn.FirstSeen,
			LastSeen:    conn.LastSeen,
			ActivePorts: maps.Clone(conn.ActivePorts),
			TotalBytes:  conn.TotalBytes,
		}
	}
	return result
}
