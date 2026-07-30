package runidentity

import (
	"crypto/sha256"
	"encoding/binary"
	"strconv"
)

func Seed(founderID string, runSeq int64) uint64 {
	digest := sha256.Sum256([]byte(founderID))
	return binary.BigEndian.Uint64(digest[:8]) ^ uint64(runSeq)
}

func SeedString(founderID string, runSeq int64) string {
	return strconv.FormatUint(Seed(founderID, runSeq), 10)
}
