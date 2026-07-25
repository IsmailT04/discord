package hasher

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/ismailtemuroglu/discord/internal/identity/ports"
	"golang.org/x/crypto/argon2"
)

const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	argonSaltLen = 16
)

// Argon2id implements ports.PasswordHasher using argon2id.
type Argon2id struct{}

// New returns an argon2id password hasher.
func New() *Argon2id {
	return &Argon2id{}
}

// Hash returns an encoded argon2id hash of password.
func (a *Argon2id) Hash(password string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return encodeHash(salt, hash), nil
}

// Compare reports whether password matches the encoded hash.
func (a *Argon2id) Compare(encoded, password string) error {
	params, salt, hash, err := decodeHash(encoded)
	if err != nil {
		return err
	}

	other := argon2.IDKey([]byte(password), salt, params.time, params.memory, params.threads, uint32(len(hash)))
	if subtle.ConstantTimeCompare(hash, other) != 1 {
		return errors.New("password mismatch")
	}
	return nil
}

type argonParams struct {
	memory  uint32
	time    uint32
	threads uint8
}

func encodeHash(salt, hash []byte) string {
	return fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argonMemory,
		argonTime,
		argonThreads,
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash),
	)
}

func decodeHash(encoded string) (argonParams, []byte, []byte, error) {
	var zero argonParams
	parts := strings.Split(encoded, "$")
	// "", "argon2id", "v=19", "m=...,t=...,p=...", salt, hash
	if len(parts) != 6 || parts[1] != "argon2id" {
		return zero, nil, nil, errors.New("invalid argon2id hash format")
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return zero, nil, nil, errors.New("invalid argon2id version")
	}
	if version != argon2.Version {
		return zero, nil, nil, fmt.Errorf("unsupported argon2 version %d", version)
	}

	var memory, timeCost uint32
	var threads int
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &timeCost, &threads); err != nil {
		return zero, nil, nil, errors.New("invalid argon2id parameters")
	}
	if threads < 1 || threads > 255 {
		return zero, nil, nil, errors.New("invalid argon2id parallelism")
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return zero, nil, nil, errors.New("invalid argon2id salt")
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return zero, nil, nil, errors.New("invalid argon2id hash")
	}

	return argonParams{memory: memory, time: timeCost, threads: uint8(threads)}, salt, hash, nil
}

var _ ports.PasswordHasher = (*Argon2id)(nil)
