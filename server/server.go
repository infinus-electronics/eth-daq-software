package server

import (
	"encoding/binary"
	"eth-daq-software/logger"
	"fmt"
	"io"
	"maps"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	// "eth-daq-software/logger"
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
	ActivePorts map[int]bool
	TotalBytes  int64
}

type DataBuffer struct {
	port           int
	clientIP       string
	buffer         []byte
	mu             sync.Mutex
	bytesReceived  int64
	lastCheck      time.Time
	rate           float64
	circularBuffer *CircularBuffer // Circular buffer to hold the last N samples
	lastAverage    float64         // Last calculated average
	leftoverByte   *byte
	hasLeftover    bool
}

// CircularBuffer implements a fixed-size circular buffer for uint16 values
type CircularBuffer struct {
	data       []float64 // Fixed-size array to hold the values
	size       int       // Total capacity of the buffer
	count      int       // Current number of elements in buffer (may be less than size)
	head       int       // Index where the next element will be inserted
	sum        float64   // Running sum of all elements in the buffer
	isFullOnce bool      // Flag indicating if the buffer has been filled at least once
}

// NewCircularBuffer creates a new circular buffer with the specified size
func NewCircularBuffer(size int) *CircularBuffer {
	return &CircularBuffer{
		data:       make([]float64, size),
		size:       size,
		count:      0,
		head:       0,
		sum:        0,
		isFullOnce: false,
	}
}

// Add adds a new value to the circular buffer, overwriting the oldest value if full
func (cb *CircularBuffer) Add(value float64) {
	// If the buffer is full, subtract the value that will be overwritten
	if cb.count == cb.size {
		// Calculate the index of the value being replaced (the oldest value)
		oldestIdx := cb.head
		cb.sum -= cb.data[oldestIdx]
	} else {
		// Buffer isn't full yet, so increment count
		cb.count++
	}

	// Add the new value to the buffer
	cb.data[cb.head] = value
	cb.sum += value

	// Move the head to the next position
	cb.head = (cb.head + 1) % cb.size

	// Mark as full once if we've reached capacity
	if cb.count == cb.size && !cb.isFullOnce {
		cb.isFullOnce = true
	}
}

// GetAverage calculates the average of all values in the buffer
func (cb *CircularBuffer) GetAverage() float64 {
	if cb.count == 0 {
		return 0.0
	}
	return float64(cb.sum) / float64(cb.count)
}

// IsFull returns true if the buffer is at capacity
func (cb *CircularBuffer) IsFull() bool {
	return cb.count == cb.size
}

// IsFullOnce returns true if the buffer has been completely filled at least once
func (cb *CircularBuffer) IsFullOnce() bool {
	return cb.isFullOnce
}

// GetCount returns the current number of elements in the buffer
func (cb *CircularBuffer) GetCount() int {
	return cb.count
}

// GetCapacity returns the total capacity of the buffer
func (cb *CircularBuffer) GetCapacity() int {
	return cb.size
}

func NewDataBuffer(port int, clientIP string, avgWindowSize int) *DataBuffer {
	return &DataBuffer{
		port:           port,
		clientIP:       SanitizeFilename(clientIP),
		buffer:         make([]byte, 0, BUFFER_SIZE),
		lastCheck:      time.Now(),
		lastAverage:    0,
		circularBuffer: NewCircularBuffer(avgWindowSize),
		leftoverByte:   nil,
		hasLeftover:    false,
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
	//handles the uint16 average calculation
	db.processBytes(data)

	elapsed := time.Since(db.lastCheck).Seconds()
	if elapsed >= 1.0 {
		rate := float64(db.bytesReceived) / elapsed / 1024 / 1024 // MB/s
		db.rate = rate
		logger.Debugf("Port %d - %s Rate: %.2f MB/s\n", db.port, db.clientIP, rate)
		db.bytesReceived = 0
		db.lastCheck = time.Now()
	}

	if len(db.buffer) >= BUFFER_SIZE {
		db.Flush()
	}
}

// processBytes converts the raw bytes to uint16 samples, handling any byte alignment issues
func (db *DataBuffer) processBytes(newBytes []byte) {
	// Start with an empty temporary buffer
	tempBuffer := make([]byte, 0, len(newBytes)+1) // +1 for potential leftover

	// If we had a leftover byte from previous data, prepend it
	if db.hasLeftover && db.leftoverByte != nil {
		tempBuffer = append(tempBuffer, *db.leftoverByte)
	}

	// Add the new bytes
	tempBuffer = append(tempBuffer, newBytes...)

	// Process all complete uint16 samples (pairs of bytes)
	completeBytes := len(tempBuffer) - (len(tempBuffer) % 2)
	for i := 0; i < completeBytes; i += 2 {
		var sample float64
		if db.port == 5555 {
			// Convert pair of bytes to uint16 (big-endian)
			// HS ADC sample processing
			sample = float64(int16(binary.LittleEndian.Uint16(tempBuffer[i : i+2])))
			sample = sample * -1 / 32768 * 2.5
		} else {
			// GADC sample processing
			sample = float64(binary.LittleEndian.Uint16(tempBuffer[i : i+2]))
			sample = sample*187.5e-6 - 6.144
		}

		// Add to our circular buffer
		db.circularBuffer.Add(sample)
	}

	// Check if we have a leftover byte
	if len(tempBuffer)%2 != 0 {
		// Store the leftover byte for the next data chunk
		leftover := tempBuffer[len(tempBuffer)-1]
		db.leftoverByte = &leftover
		db.hasLeftover = true
	} else {
		// No leftover byte
		db.leftoverByte = nil
		db.hasLeftover = false
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
			logger.Errorf("Failed to write file: %v\n", err)
		} else {
			logger.Infof("Written %d bytes to %s\n", len(data), filename)
		}
	}(data, filename)
}

func (s *Server) StartListener(port int) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		logger.Errorf("Failed to start server on port %d: %v\n", port, err)
		return
	}
	defer listener.Close()

	logger.Infof("TCP Server listening on port %d\n", port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Errorf("Failed to accept connection on port %d: %v\n", port, err)
			continue
		}

		clientIP := GetClientIP(conn.RemoteAddr())
		logger.Infof("New connection on port %d from %s\n", port, clientIP)

		buffer := NewDataBuffer(port, clientIP, 1000)
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

		logger.Infof("Connection closed from %s:%d\n", buffer.clientIP, buffer.port)
	}()

	chunk := make([]byte, 1048576) // 1MB chunks
	for {
		n, err := conn.Read(chunk)
		if err != nil {
			if err != io.EOF {
				logger.Errorf("Error reading from %s:%d: %v\n",
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

// TODO: do we really need sanitized IPs?
// AddIPConnection records or updates an IP connection
func (s *Server) AddIPConnection(ip string, port int) {
	s.connectedIPsLock.Lock()
	defer s.connectedIPsLock.Unlock()

	sanitizedIP := SanitizeFilename(ip)

	if conn, exists := s.connectedIPs[sanitizedIP]; exists {

		conn.ActivePorts[port] = true
	} else {
		s.connectedIPs[sanitizedIP] = &IPConnection{

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

		ActivePorts: maps.Clone(conn.ActivePorts),
		TotalBytes:  conn.TotalBytes,
	}, true
}

// GetAllConnectedIPs returns information about all connected IPs
func (s *Server) GetAllConnectedIPs() map[string]IPConnection {
	s.connectedIPsLock.RLock()
	defer s.connectedIPsLock.RUnlock()

	// Create a deep copy of the map
	result := make(map[string]IPConnection)
	for ip, conn := range s.connectedIPs {
		result[ip] = IPConnection{

			ActivePorts: maps.Clone(conn.ActivePorts),
			TotalBytes:  conn.TotalBytes,
		}
	}
	return result
}

func (s *Server) GetPortAverage(key BufferKey) (float64, bool) {
	s.buffersLock.RLock()
	defer s.buffersLock.RUnlock()
	// fmt.Printf("%s,%d \n", key.IP, key.Port)

	if buffer, exists := s.buffers[key]; exists {
		// fmt.Println(buffer.clientIP)
		return buffer.CalculateAverage()
	} else {
		return 0.0, false
	}
}

// CalculateAverage calculates the current average of samples in the circular buffer
// Returns the average and whether the buffer has been filled at least once
func (db *DataBuffer) CalculateAverage() (float64, bool) {
	db.mu.Lock()
	defer db.mu.Unlock()

	db.lastAverage = db.circularBuffer.GetAverage()
	isFullOnce := db.circularBuffer.IsFullOnce()

	return db.lastAverage, isFullOnce
}

// GetLastAverage returns the last calculated average without recalculating
func (db *DataBuffer) GetLastAverage() float64 {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.lastAverage
}

// GetBufferStatus returns the current state of the circular buffer (count/capacity)
func (db *DataBuffer) GetBufferStatus() (int, int) {
	db.mu.Lock()
	defer db.mu.Unlock()

	return db.circularBuffer.GetCount(), db.circularBuffer.GetCapacity()
}
