package guild

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"
)

func newGuildID(now time.Time) (string, error) {
	milliseconds := now.UnixMilli()
	if milliseconds < 0 || milliseconds >= 1<<48 {
		return "", fmt.Errorf("guild timestamp outside UUIDv7 range")
	}
	var value [16]byte
	if _, err := rand.Read(value[6:]); err != nil {
		return "", err
	}
	value[0] = byte(milliseconds >> 40)
	value[1] = byte(milliseconds >> 32)
	value[2] = byte(milliseconds >> 24)
	value[3] = byte(milliseconds >> 16)
	value[4] = byte(milliseconds >> 8)
	value[5] = byte(milliseconds)
	value[6] = 0x70 | value[6]&0x0f
	value[8] = 0x80 | value[8]&0x3f
	first := binary.BigEndian.Uint32(value[0:4])
	second := binary.BigEndian.Uint16(value[4:6])
	third := binary.BigEndian.Uint16(value[6:8])
	fourth := binary.BigEndian.Uint16(value[8:10])
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", first, second, third, fourth, value[10:]), nil
}
