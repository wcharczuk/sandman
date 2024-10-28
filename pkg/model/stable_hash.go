package model

import (
	"crypto/md5"
	"encoding/binary"
)

// StableHash implements the default hash function with
// a stable crc64 table checksum.
func StableHash(data []byte) uint64 {
	hash := md5.Sum(data)
	return binary.BigEndian.Uint64(hash[8:])
}
