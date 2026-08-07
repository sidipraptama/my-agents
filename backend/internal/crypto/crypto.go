// Package crypto provides AES-256-GCM encryption for model API keys at rest.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

type Box struct {
	gcm cipher.AEAD
}

func New(key string) (*Box, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Box{gcm: gcm}, nil
}

func (b *Box) Encrypt(plaintext string) ([]byte, error) {
	nonce := make([]byte, b.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return b.gcm.Seal(nonce, nonce, []byte(plaintext), nil), nil
}

func (b *Box) Decrypt(ciphertext []byte) (string, error) {
	nonceSize := b.gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, data := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := b.gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}
