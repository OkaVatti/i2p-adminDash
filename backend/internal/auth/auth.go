package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"

	"github.com/golang-jwt/jwt/v5"
)

var (
	// Argon2id recommended parameters (adjust to your host)
	ArgTime    uint32 = 3
	ArgMemory  uint32 = 64 * 1024
	ArgThreads uint8  = 2
	ArgKeyLen  uint32 = 32

	encryptedDBPath = "db_creds.enc"
	randReader      = rand.Reader
)

// GenerateSalt returns cryptographically strong random bytes
func GenerateSalt(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := io.ReadFull(randReader, b)
	return b, err
}

// HashToPHC produces a PHC-style argon2id encoded string and returns the raw key too
// Input `sha3hex` should be the hex string produced by client-side SHA3-512
func HashToPHC(sha3hex string) (phc string, rawKey []byte, err error) {
	if sha3hex == "" {
		return "", nil, errors.New("empty input")
	}
	salt, err := GenerateSalt(16)
	if err != nil {
		return "", nil, err
	}
	raw := argon2.IDKey([]byte(sha3hex), salt, ArgTime, ArgMemory, ArgThreads, ArgKeyLen)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(raw)
	phc = fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", ArgMemory, ArgTime, ArgThreads, b64Salt, b64Hash)
	return phc, raw, nil
}

// VerifyPHC checks sha3hex against stored PHC string
// Returns the derived raw key (for DB decryption) on success
func VerifyPHC(sha3hex string, phc string) (derived []byte, ok bool, err error) {
	// parse phc: $argon2id$v=19$m=65536,t=3,p=2$<salt>$<hash>
	parts := strings.Split(phc, "$")
	if len(parts) < 6 {
		return nil, false, errors.New("invalid PHC format")
	}
	params := parts[3]        // e.g. v=19
	_ = params                // not used
	memTimeThread := parts[4] // e.g. m=65536,t=3,p=2
	saltB64 := parts[5]
	hashB64 := ""
	if len(parts) >= 7 {
		hashB64 = parts[6]
	} else {
		return nil, false, errors.New("missing hash in PHC")
	}
	// extract m,t,p
	var m uint32
	var t uint32
	var p uint8
	// memTimeThread may include commas; split on ','
	for _, kv := range strings.Split(memTimeThread, ",") {
		if strings.HasPrefix(kv, "m=") {
			v, _ := strconv.ParseUint(strings.TrimPrefix(kv, "m="), 10, 32)
			m = uint32(v)
		}
		if strings.HasPrefix(kv, "t=") {
			v, _ := strconv.ParseUint(strings.TrimPrefix(kv, "t="), 10, 32)
			t = uint32(v)
		}
		if strings.HasPrefix(kv, "p=") {
			v, _ := strconv.ParseUint(strings.TrimPrefix(kv, "p="), 10, 8)
			p = uint8(v)
		}
	}
	if m == 0 {
		m = ArgMemory
	}
	if t == 0 {
		t = ArgTime
	}
	if p == 0 {
		p = ArgThreads
	}
	salt, err := base64.RawStdEncoding.DecodeString(saltB64)
	if err != nil {
		return nil, false, err
	}
	expected, err := base64.RawStdEncoding.DecodeString(hashB64)
	if err != nil {
		return nil, false, err
	}
	derivedKey := argon2.IDKey([]byte(sha3hex), salt, uint32(t), uint32(m), uint8(p), uint32(len(expected)))
	// constant time compare
	if len(derivedKey) != len(expected) {
		return nil, false, nil
	}
	match := subtleConstantTimeCompare(derivedKey, expected)
	return derivedKey, match, nil
}

// subtleConstantTimeCompare wraps constant time compare
func subtleConstantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

///////////////////////////////////////////////
// DB credential encryption (AES-GCM)
///////////////////////////////////////////////

// EncryptDBCreds encrypts the provided DSN using the provided 32-byte key (derived argon raw)
func EncryptDBCreds(dsn string, key []byte) error {
	if len(key) < 32 {
		return errors.New("key too short")
	}
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(randReader, nonce); err != nil {
		return err
	}
	ct := aead.Seal(nonce, nonce, []byte(dsn), nil)
	return osWriteFile(encryptedDBPath, ct, 0600)
}

// DecryptDBCreds decrypts the encrypted DB creds using the provided key and returns DSN
func DecryptDBCreds(key []byte) (string, error) {
	if len(key) < 32 {
		return "", errors.New("key too short")
	}
	data, err := osReadFile(encryptedDBPath)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key[:32])
	if err != nil {
		return "", err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSize := aead.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	plain, err := aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// EncryptedDBExists helper
func EncryptedDBExists() bool {
	_, err := osStat(encryptedDBPath)
	return err == nil
}

///////////////////////////////////////////////
// JWT helpers
///////////////////////////////////////////////

type claims struct {
	jwt.RegisteredClaims
	UserID uint `json:"uid"`
}

func CreateJWT(uid uint, secret string, ttl time.Duration) (string, error) {
	claims := &claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		UserID: uid,
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

func VerifyJWT(tokenStr, secret string) (uint, error) {
	tok, err := jwt.ParseWithClaims(tokenStr, &claims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return 0, err
	}
	if claims, ok := tok.Claims.(*claims); ok && tok.Valid {
		return claims.UserID, nil
	}
	return 0, errors.New("invalid token")
}

/////////////////////////////////////////////////
// small stdlib wrappers to allow easier testing
/////////////////////////////////////////////////

// these wrappers let tests replace with fakes if needed

var osWriteFile = func(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}
var osReadFile = func(path string) ([]byte, error) {
	return os.ReadFile(path)
}
var osStat = func(path string) (os.FileInfo, error) {
	return os.Stat(path)
}
