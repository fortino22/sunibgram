package helper

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
)

const (
	chatKeySize   = 32
	chatNonceSize = 12
)

func ChatGenerateKey() ([]byte, error) {
	key := make([]byte, chatKeySize)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return nil, err
	}
	return key, nil
}

func ChatMasterKeyFromEnv(name string) ([]byte, error) {
	value := os.Getenv(name)
	if value == "" {
		return nil, fmt.Errorf("%s is required", name)
	}
	return ChatDecodeKey(value)
}

func ChatDecodeKey(value string) ([]byte, error) {
	if decoded, err := base64.StdEncoding.DecodeString(value); err == nil && len(decoded) == chatKeySize {
		return decoded, nil
	}
	if len(value) == chatKeySize {
		return []byte(value), nil
	}
	return nil, errors.New("encryption key must be 32 raw bytes or base64-encoded 32 bytes")
}

func ChatEncrypt(plaintext, key []byte) (string, error) {
	if len(key) != chatKeySize {
		return "", errors.New("encryption key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, chatNonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(sealed), nil
}

func ChatDecrypt(encoded string, key []byte) ([]byte, error) {
	if len(key) != chatKeySize {
		return nil, errors.New("encryption key must be 32 bytes")
	}
	payload, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, err
	}
	if len(payload) < chatNonceSize {
		return nil, errors.New("ciphertext is too short")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := payload[:chatNonceSize]
	ciphertext := payload[chatNonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}
