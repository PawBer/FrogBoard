package passwords

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

func GenerateHash(password string) (string, error) {
	salt, err := genRandomBytes(16)
	if err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(password), salt, 2, 19*1024, 2, 32)

	base64Salt := base64.RawStdEncoding.EncodeToString(salt)
	base64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, 19*1024, 2, 2, base64Salt, base64Hash)

	return encodedHash, nil
}

func VerifyPassword(password, hash string) bool {
	vals := strings.Split(hash, "$")
	if len(vals) != 6 {
		return false
	}

	var version int
	_, err := fmt.Sscanf(vals[2], "v=%d", &version)
	if err != nil {
		return false
	}
	if version != argon2.Version {
		return false
	}

	var memory, iterations, parallelism int
	_, err = fmt.Sscanf(vals[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.Strict().DecodeString(vals[4])
	if err != nil {
		return false
	}

	decodedHash, err := base64.RawStdEncoding.Strict().DecodeString(vals[5])
	if err != nil {
		return false
	}

	testHash := argon2.IDKey([]byte(password), salt, uint32(iterations), uint32(memory), uint8(parallelism), uint32(len(decodedHash)))

	if subtle.ConstantTimeCompare(decodedHash, testHash) == 1 {
		return true
	} else {
		return false
	}
}

func genRandomBytes(n int) ([]byte, error) {
	buf := make([]byte, n)
	_, err := rand.Read(buf)
	if err != nil {
		return nil, err
	}

	return buf, nil
}
