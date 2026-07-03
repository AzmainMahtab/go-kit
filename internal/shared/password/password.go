// Package password defines the password hashing port and a production-grade
// Argon2id implementation.
//
// Why Argon2id?
// - Winner of the Password Hashing Competition (PHC).
// - Resistant to both GPU cracking and side-channel attacks.
// - Recommended by OWASP for password storage.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidHash         = errors.New("invalid password hash format")
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
	ErrMismatchedPassword  = errors.New("password does not match")
)

// Hashed is a tiny value object that wraps a password hash string.
// It prevents accidental confusion between plaintext and hashed passwords.
type Hashed struct{ value string }

// NewHashed creates a Hashed value object.
func NewHashed(hash string) Hashed { return Hashed{value: hash} }

func (h Hashed) String() string { return h.value }

// Hasher is the domain port. Domain code depends on this interface, not the
// Argon2 implementation.
type Hasher interface {
	Hash(password string) (string, error)
	Verify(password, encodedHash string) error
}

// Argon2idHasher implements Hasher using the Argon2id algorithm.
type Argon2idHasher struct {
	time    uint32
	memory  uint32
	threads uint8
	keyLen  uint32
	saltLen uint32
}

// NewArgon2id returns a hasher with OWASP-recommended parameters.
func NewArgon2id() *Argon2idHasher {
	return &Argon2idHasher{
		time:    3,
		memory:  64 * 1024, // 64 MB
		threads: 4,
		keyLen:  32,
		saltLen: 16,
	}
}

// Hash creates an encoded hash string from a plaintext password.
func (h *Argon2idHasher) Hash(password string) (string, error) {
	salt := make([]byte, h.saltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, h.time, h.memory, h.threads, h.keyLen)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, h.memory, h.time, h.threads, b64Salt, b64Hash,
	)
	return encodedHash, nil
}

// Verify checks a plaintext password against an encoded Argon2id hash.
func (h *Argon2idHasher) Verify(password, encodedHash string) error {
	p, salt, hash, err := decodeHash(encodedHash)
	if err != nil {
		return err
	}

	otherHash := argon2.IDKey([]byte(password), salt, p.time, p.memory, p.threads, p.keyLen)

	if subtle.ConstantTimeEq(int32(len(hash)), int32(len(otherHash))) == 0 {
		return ErrMismatchedPassword
	}
	if subtle.ConstantTimeCompare(hash, otherHash) == 1 {
		return nil
	}
	return ErrMismatchedPassword
}

type params struct {
	memory  uint32
	time    uint32
	threads uint8
	keyLen  uint32
}

func decodeHash(encodedHash string) (*params, []byte, []byte, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return nil, nil, nil, ErrInvalidHash
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return nil, nil, nil, err
	}
	if version != argon2.Version {
		return nil, nil, nil, ErrIncompatibleVersion
	}

	p := &params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.threads)
	if err != nil {
		return nil, nil, nil, err
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLen = uint32(len(parts[5]))

	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, nil, nil, err
	}
	p.keyLen = uint32(len(hash))

	return p, salt, hash, nil
}
