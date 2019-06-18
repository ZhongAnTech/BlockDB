package bytes

// GetInt32 get an int32 from byte array with a start position.
// This is for those little-endian bytes.
func GetInt32(b []byte, pos int) int32 {
	return int32(b[pos]) | int32(b[pos+1])<<8 | int32(b[pos+2])<<16 | int32(b[pos+3])<<24
}

// SetInt32 set an int32 into byte array at a position.
func SetInt32(b []byte, pos int, i int32) {
	b[pos] = byte(i)
	b[pos+1] = byte(i >> 8)
	b[pos+2] = byte(i >> 16)
	b[pos+3] = byte(i >> 24)
}
