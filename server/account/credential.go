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
	argonMemoryKiB      = 19 * 1024
	argonIterations     = 2
	argonParallelism    = 1
	argonMemoryMaxKiB   = argonMemoryKiB * 4
	argonIterationsMax  = argonIterations * 4
	argonParallelismMax = argonParallelism * 4
	argonSaltBytes      = 16
	argonKeyBytes       = 32
)

var errInvalidCredential = errors.New("invalid recovery credential")

// dummyRecoveryHash is structurally valid and deliberately never corresponds
// to a user credential. Missing-account login attempts still pay one Argon2id
// verification so account existence is not exposed by the password KDF cost.
const dummyRecoveryHash = "$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

type recoveryHashParameters struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
}

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
	code = normalizeRecoveryCode(code)
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
	valid, _ := verifyRecoveryCodeForUpgrade(encoded, code)
	return valid
}

func verifyRecoveryCodeForUpgrade(encoded, code string) (valid, needsRehash bool) {
	code = normalizeRecoveryCode(code)
	if !validRecoveryCode(code) {
		return false, false
	}
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v="+strconv.Itoa(argon2.Version) {
		return false, false
	}
	var parameters recoveryHashParameters
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &parameters.memory, &parameters.iterations, &parameters.parallelism); err != nil ||
		parts[3] != fmt.Sprintf("m=%d,t=%d,p=%d", parameters.memory, parameters.iterations, parameters.parallelism) ||
		parameters.memory < argonMemoryKiB || parameters.memory > argonMemoryMaxKiB ||
		parameters.iterations < argonIterations || parameters.iterations > argonIterationsMax ||
		parameters.parallelism < argonParallelism || parameters.parallelism > argonParallelismMax {
		return false, false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != argonSaltBytes {
		return false, false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) != argonKeyBytes {
		return false, false
	}
	got := argon2.IDKey([]byte(code), salt, parameters.iterations, parameters.memory, parameters.parallelism, uint32(len(want)))
	valid = subtle.ConstantTimeCompare(got, want) == 1
	needsRehash = valid && (parameters.memory != argonMemoryKiB || parameters.iterations != argonIterations || parameters.parallelism != argonParallelism)
	return valid, needsRehash
}

func validRecoveryCode(code string) bool {
	if len(code) != 26 || strings.ToLower(code) != code {
		return false
	}
	decoded, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(code))
	return err == nil && len(decoded) == 16
}

func normalizeRecoveryCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}
