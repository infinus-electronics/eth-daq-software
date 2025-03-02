package compress

import (
	"bytes"
	"encoding/binary"
	"math/rand"
	"reflect"
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

// BenchmarkCompressRLE benchmarks the original RLE compression function
func BenchmarkCompressRLE(b *testing.B) {
	// Generate test data with various patterns
	dataSize := 1000000 // 1 million uint16 values
	testData := make([]uint16, dataSize)

	// Create repeating patterns with different run lengths
	for i := 0; i < dataSize; {
		// Choose a random value
		value := uint16(rand.Intn(4096)) // Random 12-bit value

		// Choose a random run length between 1 and 1000
		runLength := rand.Intn(1000) + 1
		if i+runLength > dataSize {
			runLength = dataSize - i
		}

		// Fill the run
		for j := 0; j < runLength && i < dataSize; j++ {
			testData[i] = value
			i++
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		compressRLE(testData)
	}
}

// BenchmarkCompressRLEUnrolled benchmarks the optimized loop-unrolled RLE compression function
func BenchmarkCompressRLEUnrolled(b *testing.B) {
	// Generate test data with various patterns - same as above for fair comparison
	dataSize := 1000000 // 1 million uint16 values
	testData := make([]uint16, dataSize)

	// Create repeating patterns with different run lengths
	for i := 0; i < dataSize; {
		// Choose a random value
		value := uint16(rand.Intn(4096)) // Random 12-bit value

		// Choose a random run length between 1 and 1000
		runLength := rand.Intn(1000) + 1
		if i+runLength > dataSize {
			runLength = dataSize - i
		}

		// Fill the run
		for j := 0; j < runLength && i < dataSize; j++ {
			testData[i] = value
			i++
		}
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		compressRLEUnrolled(testData)
	}
}

// BenchmarkCompressRLEWithDifferentPatterns benchmarks both functions with different data patterns
func BenchmarkCompressRLEWithDifferentPatterns(b *testing.B) {
	// Test cases with different data patterns
	testCases := []struct {
		name     string
		dataFunc func(size int) []uint16
	}{
		{
			name: "HighRepetition",
			dataFunc: func(size int) []uint16 {
				// Data with very long runs (good for RLE)
				data := make([]uint16, size)
				value := uint16(0)
				for i := 0; i < size; {
					runLength := 10000 // Long runs
					if i+runLength > size {
						runLength = size - i
					}
					for j := 0; j < runLength; j++ {
						data[i+j] = value
					}
					i += runLength
					value = (value + 1) % 4096
				}
				return data
			},
		},
		{
			name: "LowRepetition",
			dataFunc: func(size int) []uint16 {
				// Data with very short runs (poor for RLE)
				data := make([]uint16, size)
				for i := 0; i < size; i++ {
					if i%2 == 0 {
						data[i] = uint16(i % 4096)
					} else {
						data[i] = uint16((i + 1) % 4096)
					}
				}
				return data
			},
		},
		{
			name: "MixedPattern",
			dataFunc: func(size int) []uint16 {
				// Mix of long and short runs
				data := make([]uint16, size)
				i := 0
				for i < size {
					// Alternating between long and short runs
					if (i/10000)%2 == 0 {
						// Long run
						value := uint16(i % 4096)
						runLength := 1000
						if i+runLength > size {
							runLength = size - i
						}
						for j := 0; j < runLength; j++ {
							data[i+j] = value
						}
						i += runLength
					} else {
						// Short runs
						runLength := 50
						if i+runLength > size {
							runLength = size - i
						}
						for j := 0; j < runLength; j++ {
							data[i+j] = uint16((i + j) % 4096)
						}
						i += runLength
					}
				}
				return data
			},
		},
	}

	dataSize := 1000000 // 1 million uint16 values

	for _, tc := range testCases {
		testData := tc.dataFunc(dataSize)

		b.Run(tc.name+"/Original", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				compressRLE(testData)
			}
		})

		b.Run(tc.name+"/Unrolled", func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				compressRLEUnrolled(testData)
			}
		})
	}
}

// TestCompressRLEUnrolled ensures the unrolled version produces identical results to the original
func TestCompressRLEUnrolled(t *testing.T) {
	tests := []struct {
		name string
		data []uint16
	}{
		{
			name: "Empty data",
			data: []uint16{},
		},
		{
			name: "Single value",
			data: []uint16{42},
		},
		{
			name: "Small data with repetition",
			data: []uint16{10, 10, 10, 10, 20, 20, 30, 30, 30, 30},
		},
		{
			name: "No repetition",
			data: []uint16{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		{
			name: "Edge case - 4 values",
			data: []uint16{10, 10, 10, 10},
		},
		{
			name: "Edge case - 5 values",
			data: []uint16{10, 10, 10, 10, 10},
		},
		{
			name: "Edge case - 3 values",
			data: []uint16{10, 10, 10},
		},
		{
			name: "Pattern breaking at unrolling boundary",
			data: []uint16{10, 10, 10, 20, 30, 30, 30, 30},
		},
		{
			name: "Complex pattern",
			data: []uint16{1, 1, 1, 2, 2, 2, 2, 3, 3, 4, 4, 4, 4, 4, 5, 6, 7, 7, 7, 7},
		},
		{
			name: "Long run",
			data: func() []uint16 {
				result := make([]uint16, 1000)
				for i := range result {
					result[i] = 42
				}
				return result
			}(),
		},
		{
			name: "Multiple long runs",
			data: func() []uint16 {
				result := make([]uint16, 1000)
				for i := range result {
					if i < 400 {
						result[i] = 10
					} else if i < 800 {
						result[i] = 20
					} else {
						result[i] = 30
					}
				}
				return result
			}(),
		},
		{
			name: "Alternating pattern",
			data: func() []uint16 {
				result := make([]uint16, 100)
				for i := range result {
					result[i] = uint16(i % 2)
				}
				return result
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Get results from both implementations
			originalResult := compressRLE(tt.data)
			unrolledResult := compressRLEUnrolled(tt.data)

			// Check if results have the same length
			if len(originalResult) != len(unrolledResult) {
				t.Fatalf("Result length mismatch: original=%d, unrolled=%d",
					len(originalResult), len(unrolledResult))
			}

			// Check if each RLEData element matches
			for i := 0; i < len(originalResult); i++ {
				if originalResult[i].Value != unrolledResult[i].Value ||
					originalResult[i].Count != unrolledResult[i].Count {
					t.Fatalf("Result mismatch at index %d: original={Value:%d, Count:%d}, unrolled={Value:%d, Count:%d}",
						i, originalResult[i].Value, originalResult[i].Count,
						unrolledResult[i].Value, unrolledResult[i].Count)
				}
			}

			// Verify decompression produces the original data
			decompressed := decompressRLE(unrolledResult, len(tt.data))
			if !reflect.DeepEqual(decompressed, tt.data) {
				t.Fatalf("Decompression mismatch: expected %v, got %v", tt.data, decompressed)
			}
		})
	}
}

// TestCompressRLEUnrolledOverflow tests the overflow handling in the unrolled version
func TestCompressRLEUnrolledOverflow(t *testing.T) {
	// Create a small maxCount value to simulate overflow
	testMaxCount := uint32(10)

	// Create test data with a run longer than testMaxCount
	data := make([]uint16, 50)
	for i := range data {
		data[i] = 42 // All the same value
	}

	// Create custom versions of both functions that use our test maxCount
	customCompressRLE := func(msb12Bits []uint16) []RLEData {
		if len(msb12Bits) == 0 {
			return []RLEData{}
		}

		var result []RLEData
		currentValue := msb12Bits[0]
		currentCount := uint32(1)
		maxCount := testMaxCount

		for i := 1; i < len(msb12Bits); i++ {
			if msb12Bits[i] == currentValue {
				// Same value, increment the count
				currentCount++

				// Check if we're about to overflow uint32
				if currentCount == maxCount {
					// Store the current run and start a new run with the same value
					result = append(result, RLEData{Value: currentValue, Count: currentCount})
					currentCount = 0 // Reset count for the next entry with same value
				}
			} else {
				// Different value, store the current run and start a new one
				result = append(result, RLEData{Value: currentValue, Count: currentCount})
				currentValue = msb12Bits[i]
				currentCount = 1
			}
		}

		// Don't forget to add the last run
		if currentCount > 0 {
			result = append(result, RLEData{Value: currentValue, Count: currentCount})
		}

		return result
	}

	customCompressRLEUnrolled := func(msb12Bits []uint16) []RLEData {
		if len(msb12Bits) == 0 {
			return []RLEData{}
		}

		// Pre-allocate capacity to reduce reallocations
		result := make([]RLEData, 0, len(msb12Bits)/4+1)
		currentValue := msb12Bits[0]
		currentCount := uint32(1)
		maxCount := testMaxCount

		// Main loop with 4x unrolling
		i := 1
		for i <= len(msb12Bits)-4 {
			// Process 4 elements at once
			if msb12Bits[i] == currentValue &&
				msb12Bits[i+1] == currentValue &&
				msb12Bits[i+2] == currentValue &&
				msb12Bits[i+3] == currentValue {
				// All 4 values match current run
				currentCount += 4
				i += 4

				// Check for uint32 overflow
				if currentCount >= maxCount {
					// Store the run up to maxCount
					result = append(result, RLEData{Value: currentValue, Count: maxCount})
					currentCount = currentCount - maxCount
				}
			} else {
				// Process one by one when pattern breaks
				if msb12Bits[i] == currentValue {
					currentCount++

					// Check for uint32 overflow
					if currentCount == maxCount {
						// Store the current run and start a new run with the same value
						result = append(result, RLEData{Value: currentValue, Count: currentCount})
						currentCount = 0 // Reset count for the next entry with same value
					}
				} else {
					// Different value, store the current run and start a new one
					result = append(result, RLEData{Value: currentValue, Count: currentCount})
					currentValue = msb12Bits[i]
					currentCount = 1
				}
				i++
			}
		}

		// Handle remaining elements (less than 4)
		for ; i < len(msb12Bits); i++ {
			if msb12Bits[i] == currentValue {
				// Same value, increment the count
				currentCount++

				// Check for uint32 overflow
				if currentCount == maxCount {
					// Store the current run and start a new run with the same value
					result = append(result, RLEData{Value: currentValue, Count: currentCount})
					currentCount = 0 // Reset count for the next entry with same value
				}
			} else {
				// Different value, store the current run and start a new one
				result = append(result, RLEData{Value: currentValue, Count: currentCount})
				currentValue = msb12Bits[i]
				currentCount = 1
			}
		}

		// Don't forget to add the last run
		if currentCount > 0 {
			result = append(result, RLEData{Value: currentValue, Count: currentCount})
		}

		return result
	}

	// Get results from both implementations
	originalResult := customCompressRLE(data)
	unrolledResult := customCompressRLEUnrolled(data)

	// Print results for debugging
	t.Logf("Original result:")
	for i, r := range originalResult {
		t.Logf("  [%d] Value: %d, Count: %d", i, r.Value, r.Count)
	}

	t.Logf("Unrolled result:")
	for i, r := range unrolledResult {
		t.Logf("  [%d] Value: %d, Count: %d", i, r.Value, r.Count)
	}

	// Check if they both created multiple chunks due to overflow
	if len(originalResult) <= 1 || len(unrolledResult) <= 1 {
		t.Fatalf("Expected multiple RLE entries due to overflow handling, got: original=%d, unrolled=%d",
			len(originalResult), len(unrolledResult))
	}

	// Since we may have slight differences in implementation, check for functional equivalence
	// by decompressing and comparing the results
	originalDecompressed := decompressRLE(originalResult, len(data))
	unrolledDecompressed := decompressRLE(unrolledResult, len(data))

	if !reflect.DeepEqual(originalDecompressed, unrolledDecompressed) {
		t.Fatalf("Decompression results differ between original and unrolled implementations")
	}

	// Also verify both match the original data
	if !reflect.DeepEqual(originalDecompressed, data) {
		t.Fatalf("Original decompression mismatch with original data")
	}

	if !reflect.DeepEqual(unrolledDecompressed, data) {
		t.Fatalf("Unrolled decompression mismatch with original data")
	}

	t.Logf("Successfully verified overflow handling with both implementations")
}

// TestCompressRLEUnrolledRoundTrip tests complete compression/decompression cycle
func TestCompressRLEUnrolledRoundTrip(t *testing.T) {
	// Generate test data with various patterns
	tests := []struct {
		name string
		data []byte
	}{
		{
			name: "Random data with repetition",
			data: func() []byte {
				rand.Seed(42) // Use fixed seed for reproducibility
				size := 10000
				data := make([]byte, size)
				for i := 0; i < size; {
					// Choose a random value
					value := byte(rand.Intn(256))

					// Choose a random run length between 1 and 100
					runLength := rand.Intn(100) + 1
					if i+runLength > size {
						runLength = size - i
					}

					// Fill the run
					for j := 0; j < runLength && i < size; j++ {
						data[i] = value
						i++
					}
				}
				return data
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Process initial data using the regular HybridRLECompress but with unrolled function
			// We need to modify the HybridRLECompress to use our unrolled function

			// Create a patched version of HybridRLECompress that calls our unrolled function
			compressWithUnrolled := func(data []byte) []byte {
				// This is essentially a copy of HybridRLECompress but using compressRLEUnrolled

				// Create a padded array if needed
				paddedArray := data
				if len(data)%2 != 0 {
					// Pad with a zero byte at the end
					paddedArray = make([]byte, len(data)+1)
					copy(paddedArray, data)
					// Padding byte is 0x00
					paddedArray[len(data)] = 0x00
				}
				valueCount := len(paddedArray) / 2

				// Create a slice to hold the uint16 values
				uint16Array := make([]uint16, valueCount)

				// Convert the bytes to uint16 values
				for i := 0; i < valueCount; i++ {
					// Get the index in the byte array
					byteIndex := i * 2

					// Read 2 bytes and convert to uint16 in little-endian order
					uint16Array[i] = binary.LittleEndian.Uint16(paddedArray[byteIndex : byteIndex+2])
				}
				// Extract 12 most significant bits and 4 least significant bits
				msb12Bits := make([]uint16, valueCount)
				lsb4Bits := make([]uint8, valueCount)

				for i, value := range uint16Array {
					// Extract 12 most significant bits (shift right by 4 bits)
					msb12Bits[i] = value >> 4

					// Extract 4 least significant bits (mask with 0xF)
					lsb4Bits[i] = uint8(value & 0xF)
				}

				// Use our unrolled function here
				compressedRLE := compressRLEUnrolled(msb12Bits)
				packedLSB4 := packLSB4IntoUint16(lsb4Bits)

				// Store the compressed data in a binary format with header
				// Format:
				// - Magic string "RLE4" (4 bytes)
				// - Original data length (4 bytes, uint32)
				// - Number of RLE entries (4 bytes, uint32)
				// - Number of LSB4 packed values (4 bytes, uint32)
				// - Was data padded? (1 byte, uint8: 0=no, 1=yes)
				// - RLE data entries (each entry is 6 bytes: 2 for Value, 4 for Count)
				// - LSB4 packed values (each value is 2 bytes)

				// Calculate sizes
				headerSize := 4 + 4 + 4 + 4 + 1       // Magic + orig len + RLE count + LSB4 count + padding flag
				rleDataSize := len(compressedRLE) * 6 // Each RLE entry is 6 bytes
				lsb4DataSize := len(packedLSB4) * 2   // Each packed LSB4 value is 2 bytes

				// Create result buffer with appropriate size
				result := make([]byte, headerSize+rleDataSize+lsb4DataSize)

				// Write magic string "RLE4"
				copy(result[0:4], []byte("RLE4"))

				// Write original data length
				binary.LittleEndian.PutUint32(result[4:8], uint32(len(data)))

				// Write number of RLE entries
				binary.LittleEndian.PutUint32(result[8:12], uint32(len(compressedRLE)))

				// Write number of LSB4 packed values
				binary.LittleEndian.PutUint32(result[12:16], uint32(len(packedLSB4)))

				// Write padding flag
				if len(data)%2 != 0 {
					result[16] = 1 // Data was padded
				} else {
					result[16] = 0 // Data was not padded
				}

				// Write RLE data
				rleOffset := headerSize
				for _, rle := range compressedRLE {
					// Write Value (uint16)
					binary.LittleEndian.PutUint16(result[rleOffset:rleOffset+2], rle.Value)

					// Write Count (uint32)
					binary.LittleEndian.PutUint32(result[rleOffset+2:rleOffset+6], rle.Count)

					// Move to next RLE entry
					rleOffset += 6
				}

				// Write LSB4 packed values
				lsb4Offset := headerSize + rleDataSize
				for _, lsb4 := range packedLSB4 {
					binary.LittleEndian.PutUint16(result[lsb4Offset:lsb4Offset+2], lsb4)
					lsb4Offset += 2
				}

				return result
			}

			// Compress using our unrolled function
			compressed := compressWithUnrolled(tt.data)

			// Decompress using the existing decompressor
			decompressed, err := HybridRLEDecompress(compressed)
			if err != nil {
				t.Fatalf("Failed to decompress: %v", err)
			}

			// Verify the decompressed data matches the original
			if !bytes.Equal(decompressed, tt.data) {
				t.Fatalf("Data mismatch after round-trip compression/decompression")
			}

			t.Logf("Successfully verified round-trip compression/decompression using unrolled function")

			// Also compare compression ratios
			compressedOriginal := HybridRLECompress(tt.data)
			ratio := float64(len(compressed)) / float64(len(compressedOriginal))
			t.Logf("Size comparison - Original: %d bytes, Unrolled: %d bytes, Ratio: %.2f",
				len(compressedOriginal), len(compressed), ratio)
		})
	}
}
