package account

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	argonMemoryKiB   = 19 * 1024
	argonIterations  = 2
	argonParallelism = 1
	argonSaltBytes   = 16
	argonKeyBytes    = 32
)

var errInvalidCredential = errors.New("invalid recovery credential")

func newRecoveryCode(random io.Reader) (string, error) {
	if random == nil {
		random = rand.Reader
	}
	data := make([]byte, 16)
	if _, err := io.ReadFull(random, data); err != nil {
		return "", err
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(data)), nil
}

func hashRecoveryCode(code string, random io.Reader) (string, error) {
	if random == nil {
		random = rand.Reader
	}
	if !validRecoveryCode(code) {
		return "", errInvalidCredential
	}
	salt := make([]byte, argonSaltBytes)
	if _, err := io.ReadFull(random, salt); err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(code), salt, argonIterations, argonMemoryKiB, argonParallelism, argonKeyBytes)
	encoding := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, argonMemoryKiB,
		argonIterations, argonParallelism, encoding.EncodeToString(salt), encoding.EncodeToString(hash)), nil
}

func verifyRecoveryCode(encoded, code string) bool {
	if !validRecoveryCode(code) {
		return false
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v="+strconv.Itoa(argon2.Version) {
		return false
	}
	var memory uint32
	var iterations uint32
	var parallelism uint8
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism); err != nil ||
		memory != argonMemoryKiB || iterations != argonIterations || parallelism != argonParallelism {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != argonSaltBytes {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) != argonKeyBytes {
		return false
	}
	got := argon2.IDKey([]byte(code), salt, iterations, memory, parallelism, uint32(len(want)))
	return subtle.ConstantTimeCompare(got, want) == 1
}

func validRecoveryCode(code string) bool {
	if len(code) != 26 || strings.ToLower(code) != code {
		return false
	}
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(code))
	return err == nil && len(decoded) == 16
}
