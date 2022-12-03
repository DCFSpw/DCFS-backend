package checksum

import (
	"encoding/hex"
	"golang.org/x/crypto/sha3"
)

// CalculateChecksum - calculate checksum of provided data using SHA3-256 algorithm
//
// params:
//   - data []uint8: data to calculate checksum for
//
// return type:
//   - string: checksum of provided data
func CalculateChecksum(data []uint8) string {
	hash := sha3.New256()

	if _, err := hash.Write(data); err != nil {
		return ""
	}

	return hex.EncodeToString(hash.Sum(nil))
}
