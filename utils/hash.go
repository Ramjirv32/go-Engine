package utils

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
)

// GenerateDataHash creates an MD5 hash of the college data for comparison
func GenerateDataHash(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return ""
	}

	hash := md5.Sum(jsonData)
	return fmt.Sprintf("%x", hash)
}

// CompareHashes checks if two data hashes are the same
func CompareHashes(hash1, hash2 string) bool {
	return hash1 == hash2
}
