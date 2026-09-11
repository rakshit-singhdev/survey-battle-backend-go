package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 3
	argonMemory  = 64 * 1024 // 64 MB
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

func HashPassword(password string) (string, error) {
	salt := make([]byte, argonSaltLen)

	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argonTime,
		argonMemory,
		argonThreads,
		argonKeyLen,
	)

	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf(
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		argonMemory,
		argonTime,
		argonThreads,
		encodedSalt,
		encodedHash,
	), nil
}

func ComparePassword(password, encodedHash string) bool {
	// Parse the encoded Argon2id hash
	var memory, time uint32
	var threads uint8
	var encodedSalt, encodedPasswordHash string

	_, err := fmt.Sscanf(
		encodedHash,
		"$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		&memory,
		&time,
		&threads,
		&encodedSalt,
		&encodedPasswordHash,
	)

	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(encodedSalt)
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(encodedPasswordHash)
	if err != nil {
		return false
	}

	actualHash := argon2.IDKey(
		[]byte(password),
		salt,
		time,
		memory,
		threads,
		uint32(len(expectedHash)),
	)

	if len(actualHash) != len(expectedHash) {
		return false
	}

	var result byte
	for i := range actualHash {
		result |= actualHash[i] ^ expectedHash[i]
	}

	return result == 0
}