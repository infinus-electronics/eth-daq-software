package compress

import (
	"bytes"
	"math/rand"
	"testing"
)

// TestHybridRLECompressDecompress tests the round-trip compression and decompression
func TestHybridRLECompressDecompress(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Empty data",
			data: []byte{},
		},
		{
			name: "Single byte",
			data: []byte{42},
		},
		{
			name: "Small data with repetition",
			data: []byte{1, 1, 1, 1, 2, 2, 3, 3, 3, 3},
		},
		{
			name: "Even length data",
			data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name: "Odd length data",
			data: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		{
			name: "High repetition data",
			data: bytes.Repeat([]byte{0xAA, 0xBB}, 100),
		},
		{
			name: "Low repetition data",
			data: func() []byte {
				result := make([]byte, 100)
				for i := range result {
					result[i] = byte(i % 256)
				}
				return result
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress
			compressed := HybridRLECompress(tt.data)

			// Log compression ratio if data is not empty
			if len(tt.data) > 0 {
				ratio := float64(len(compressed)) / float64(len(tt.data))
				t.Logf("Compression ratio: %.2f (original: %d bytes, compressed: %d bytes)",
					ratio, len(tt.data), len(compressed))
			}

			// Decompress
			decompressed, err := HybridRLEDecompress(compressed)
			if err != nil {
				t.Fatalf("Failed to decompress: %v", err)
			}

			// Verify
			if !bytes.Equal(decompressed, tt.data) {
				t.Fatalf("Data mismatch after decompression\nOriginal: %v\nDecompressed: %v",
					tt.data, decompressed)
			}
		})
	}
}

// TestLargeData tests compression/decompression with a larger dataset
func TestLargeData(t *testing.T) {
	// Generate 100KB of random data
	size := 100 * 1024
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(rand.Intn(256))
	}

	// Compress
	compressed := HybridRLECompress(data)
	ratio := float64(len(compressed)) / float64(len(data))
	t.Logf("Large random data compression ratio: %.2f (original: %d bytes, compressed: %d bytes)",
		ratio, len(data), len(compressed))

	// Decompress
	decompressed, err := HybridRLEDecompress(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress large data: %v", err)
	}

	// Verify
	if !bytes.Equal(decompressed, data) {
		t.Fatal("Data mismatch after decompressing large data")
	}
}

// TestHighRepetitionData tests compression with highly repetitive data
func TestHighRepetitionData(t *testing.T) {
	// Generate 100KB of highly repetitive data (good for RLE)
	size := 100 * 1024
	data := make([]byte, size)

	// Create patterns of repetition
	for i := 0; i < size; i += 100 {
		// Fill a block with the same value
		blockSize := min(100, size-i)
		value := byte(i / 100)
		for j := 0; j < blockSize; j++ {
			data[i+j] = value
		}
	}

	// Compress
	compressed := HybridRLECompress(data)
	ratio := float64(len(compressed)) / float64(len(data))
	t.Logf("High repetition data compression ratio: %.2f (original: %d bytes, compressed: %d bytes)",
		ratio, len(data), len(compressed))

	// Decompress
	decompressed, err := HybridRLEDecompress(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress high repetition data: %v", err)
	}

	// Verify
	if !bytes.Equal(decompressed, data) {
		t.Fatal("Data mismatch after decompressing high repetition data")
	}
}

// TestInvalidCompressedData tests handling of invalid compressed data
func TestInvalidCompressedData(t *testing.T) {
	tests := []struct {
		name        string
		invalidData []byte
	}{
		{
			name:        "Too short",
			invalidData: []byte{1, 2, 3},
		},
		{
			name:        "Wrong magic string",
			invalidData: append([]byte("WRONG"), make([]byte, 13)...),
		},
		{
			name: "Incorrect length",
			invalidData: func() []byte {
				// Create valid header but with incorrect data length
				data := make([]byte, 30)
				copy(data[0:4], []byte("RLE4"))
				// Set original length to 1000
				for i := 4; i < 8; i++ {
					data[i] = 0
				}
				data[4] = 232 // 1000 in little endian (first byte)
				data[5] = 3   // 1000 in little endian (second byte)
				return data
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HybridRLEDecompress(tt.invalidData)
			if err == nil {
				t.Fatal("Expected error when decompressing invalid data, got nil")
			}
			t.Logf("Got expected error: %v", err)
		})
	}
}

// TestOverflowHandling tests handling of data with potential uint32 overflow in RLE
func TestOverflowHandling(t *testing.T) {
	// Create a large array with the same value repeated many times
	// to test the overflow handling in compressRLE
	value := uint16(0xABC) // 12-bit value
	count := 5000000       // Large enough to potentially cause uint32 overflow concerns

	// Create the data
	data := make([]byte, count*2) // Each uint16 is 2 bytes
	for i := 0; i < count; i++ {
		data[i*2] = byte(value & 0xFF)
		data[i*2+1] = byte(value >> 8)
	}

	// Compress
	compressed := HybridRLECompress(data)

	// Decompress
	decompressed, err := HybridRLEDecompress(compressed)
	if err != nil {
		t.Fatalf("Failed to decompress overflow test data: %v", err)
	}

	// Verify
	if !bytes.Equal(decompressed, data) {
		t.Fatal("Data mismatch after decompressing overflow test data")
	}

	t.Logf("Successfully handled large repetition count (%d)", count)
}

// Helper function min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Benchmark compression and decompression
func BenchmarkHybridRLECompression(b *testing.B) {
	// Generate 10MB of data with moderate repetition
	size := 10 * 1024 * 1024
	data := make([]byte, size)

	// Create some repetition patterns
	for i := 0; i < size; i++ {
		// Create areas of repetition and areas of randomness
		if (i/1000)%2 == 0 {
			// Repeated areas
			data[i] = byte((i / 1000) % 256)
		} else {
			// Random areas
			data[i] = byte(i % 256)
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		HybridRLECompress(data)
	}
}

func BenchmarkHybridRLEDecompression(b *testing.B) {
	// Generate and compress 10MB of data
	size := 10 * 1024 * 1024
	data := make([]byte, size)
	for i := 0; i < size; i++ {
		if (i/1000)%2 == 0 {
			data[i] = byte((i / 1000) % 256)
		} else {
			data[i] = byte(i % 256)
		}
	}

	compressed := HybridRLECompress(data)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := HybridRLEDecompress(compressed)
		if err != nil {
			b.Fatalf("Decompression failed: %v", err)
		}
	}
}
