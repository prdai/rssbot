// Package utils is for helper and utility related work
package utils

import (
	"crypto/sha256"
	"encoding/binary"
	"log/slog"
	"strconv"

	"github.com/mitchellh/hashstructure/v2"
)

func ConvertStringToHash(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	hashBytes := h.Sum(nil)
	return strconv.FormatUint(binary.LittleEndian.Uint64(hashBytes), 2)
}

func ConvertObjectToHash(obj any) (string, error) {
	hash, err := hashstructure.Hash(obj, hashstructure.FormatV2, nil)
	if err != nil {
		slog.Error(err.Error())
		return *new(string), err
	}
	return strconv.FormatUint(hash, 2), nil
}
