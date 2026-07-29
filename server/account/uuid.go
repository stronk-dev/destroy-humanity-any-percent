package account

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

func newUUIDv7(now time.Time, random io.Reader) (string, error) {
	if random == nil {
		random = rand.Reader
	}
	var value [16]byte
	if _, err := io.ReadFull(random, value[:]); err != nil {
		return "", err
	}
	milliseconds := now.UTC().UnixMilli()
	if milliseconds < 0 || uint64(milliseconds) >= 1<<48 {
		return "", fmt.Errorf("timestamp outside UUIDv7 range")
	}
	var timestamp [8]byte
	binary.BigEndian.PutUint64(timestamp[:], uint64(milliseconds))
	copy(value[0:6], timestamp[2:])
	value[6] = value[6]&0x0f | 0x70
	value[8] = value[8]&0x3f | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
