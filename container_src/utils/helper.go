package utils

import (
	"crypto/sha256"
	"encoding/binary"
)

func ConvertStringToHash(str string) uint64 {
	h := sha256.New()
	h.Write([]byte(str))
	hashBytes := h.Sum(nil)
	return binary.LittleEndian.Uint64(hashBytes)
}
