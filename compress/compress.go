package compress

import (
	"encoding/binary"
	"fmt"
)

// RLEData stores run-length encoded data
type RLEData struct {
	Value uint16 // The value that is repeated
	Count uint32 // How many times it repeats
}

// compressRLE compresses the 12 MSB array using run-length encoding
// Handles cases where run length might exceed uint32 max value
func compressRLE(msb12Bits []uint16) []RLEData {
	if len(msb12Bits) == 0 {
		return []RLEData{}
	}

	var result []RLEData
	currentValue := msb12Bits[0]
	currentCount := uint32(1)
	maxCount := uint32(^uint32(0)) // Maximum value of uint32

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
	result = append(result, RLEData{Value: currentValue, Count: currentCount})

	// sort.Slice(result, func(i, j int) bool {
	// 	return result[i].Count > result[j].Count
	// })

	return result
}

// decompressRLE decompresses RLE data back to the original MSB12 array
func decompressRLE(compressedData []RLEData, expectedLength int) []uint16 {
	result := make([]uint16, 0, expectedLength)

	for _, rle := range compressedData {
		// For very large count values, use a more efficient approach
		// than appending individual values in a loop
		if rle.Count > 1000 {
			// Create a slice of the same value
			valueSlice := make([]uint16, rle.Count)
			for i := range valueSlice {
				valueSlice[i] = rle.Value
			}
			result = append(result, valueSlice...)
		} else {
			// For smaller counts, a simple loop is fine
			for i := uint32(0); i < rle.Count; i++ {
				result = append(result, rle.Value)
			}
		}
	}

	return result
}

// compressRLEUnrolled compresses the 12 MSB array using run-length encoding with loop unrolling
// Handles cases where run length might exceed uint32 max value
func compressRLEUnrolled(msb12Bits []uint16) []RLEData {
	if len(msb12Bits) == 0 {
		return []RLEData{}
	}

	// Pre-allocate capacity to reduce reallocations
	result := make([]RLEData, 0, len(msb12Bits)/4+1)
	currentValue := msb12Bits[0]
	currentCount := uint32(1)
	maxCount := uint32(^uint32(0)) // Maximum value of uint32

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
			if currentCount == maxCount {
				// Store the current run and start a new run with the same value
				result = append(result, RLEData{Value: currentValue, Count: currentCount})
				currentCount = 0 // Reset count for the next entry with same value
			}
		} else {
			// Process one element at a time when pattern breaks
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

// packLSB4IntoUint16 packs four 4-bit values into each uint16
func packLSB4IntoUint16(lsb4Bits []uint8) []uint16 {
	// Calculate how many uint16 values we need
	// Each uint16 can hold 4 LSB4 values (4 bits each)
	numValues := (len(lsb4Bits) + 3) / 4 // Ceiling division

	// Create the result array
	packedValues := make([]uint16, numValues)

	// Pack 4 LSB4 values into each uint16
	for i := 0; i < len(lsb4Bits); i++ {
		// Which uint16 this value belongs to
		packedIndex := i / 4

		// Which position within the uint16 (0-3)
		// 0 is least significant, 3 is most significant
		positionInUint16 := i % 4

		// Shift the 4-bit value to its position and OR it into the result
		// Position 0: bits 0-3, Position 1: bits 4-7, Position 2: bits 8-11, Position 3: bits 12-15
		shiftAmount := positionInUint16 * 4
		packedValues[packedIndex] |= uint16(lsb4Bits[i]&0xF) << shiftAmount
	}

	return packedValues
}

// unpackUint16ToLSB4 unpacks uint16 values back to 4-bit values
func unpackUint16ToLSB4(packedValues []uint16, originalLength int) []uint8 {
	// Create the result array
	unpackedValues := make([]uint8, originalLength)

	// Unpack each uint16 into 4 LSB4 values
	for i := 0; i < originalLength; i++ {
		// Which uint16 this value comes from
		packedIndex := i / 4

		// Which position within the uint16 (0-3)
		positionInUint16 := i % 4

		// Extract the 4-bit value from its position
		shiftAmount := positionInUint16 * 4
		unpackedValues[i] = uint8((packedValues[packedIndex] >> shiftAmount) & 0xF)
	}

	return unpackedValues
}

func HybridRLECompress(data []byte) []byte {
	// Create a padded array if needed
	// fmt.Println(len(data))
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

	compressedRLE := compressRLE(msb12Bits)
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

// HybridRLEDecompress decompresses data that was compressed with HybridRLECompress
func HybridRLEDecompress(compressedData []byte) ([]byte, error) {
	// Check if there's enough data for the header
	if len(compressedData) < 17 {
		return nil, fmt.Errorf("compressed data too short to contain valid header")
	}

	// Check magic string
	if string(compressedData[0:4]) != "RLE4" {
		return nil, fmt.Errorf("invalid magic string: expected 'RLE4'")
	}

	// Read header data
	originalLength := binary.LittleEndian.Uint32(compressedData[4:8])
	rleEntryCount := binary.LittleEndian.Uint32(compressedData[8:12])
	lsb4Count := binary.LittleEndian.Uint32(compressedData[12:16])
	wasPadded := compressedData[16] == 1

	// Calculate offsets
	headerSize := 17
	rleDataSize := int(rleEntryCount) * 6
	rleOffset := headerSize
	lsb4Offset := headerSize + rleDataSize

	// Ensure the compressed data contains all expected sections
	if len(compressedData) < headerSize+rleDataSize+int(lsb4Count)*2 {
		return nil, fmt.Errorf("compressed data is too short to contain all expected sections")
	}

	// Read RLE data
	compressedRLE := make([]RLEData, rleEntryCount)
	for i := 0; i < int(rleEntryCount); i++ {
		// Read Value (uint16)
		value := binary.LittleEndian.Uint16(compressedData[rleOffset : rleOffset+2])

		// Read Count (uint32)
		count := binary.LittleEndian.Uint32(compressedData[rleOffset+2 : rleOffset+6])

		compressedRLE[i] = RLEData{
			Value: value,
			Count: count,
		}

		// Move to next RLE entry
		rleOffset += 6
	}

	// Read LSB4 packed values
	packedLSB4 := make([]uint16, lsb4Count)
	for i := 0; i < int(lsb4Count); i++ {
		packedLSB4[i] = binary.LittleEndian.Uint16(compressedData[lsb4Offset : lsb4Offset+2])
		lsb4Offset += 2
	}

	// Calculate how many values we expect after decompression
	// This should match the number of values we compressed
	valueCount := 0
	for _, rle := range compressedRLE {
		valueCount += int(rle.Count)
	}

	// Decompress RLE data
	msb12Bits := decompressRLE(compressedRLE, valueCount)

	// Unpack LSB4 values
	lsb4Bits := unpackUint16ToLSB4(packedLSB4, valueCount)

	// Combine MSB12 and LSB4 back into original uint16 values
	uint16Array := make([]uint16, valueCount)
	for i := 0; i < valueCount; i++ {
		msb12 := msb12Bits[i]
		lsb4 := uint16(lsb4Bits[i])
		uint16Array[i] = (msb12 << 4) | lsb4
	}

	// Convert uint16 values back to bytes
	result := make([]byte, valueCount*2)
	for i, value := range uint16Array {
		byteIndex := i * 2
		binary.LittleEndian.PutUint16(result[byteIndex:byteIndex+2], value)
	}

	// If the original data was padded, remove the padding
	if wasPadded && len(result) > 0 {
		result = result[:len(result)-1]
	}

	// Verify the length
	if uint32(len(result)) != originalLength {
		return nil, fmt.Errorf("decompressed data length (%d) doesn't match expected length (%d)", len(result), originalLength)
	}

	return result, nil
}
