package cipher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"math"
)

type Cipher struct {
	Key []byte
}

func NewCipher(key []byte) *Cipher {
	return &Cipher{
		Key: padKey(key),
	}
}

func (c *Cipher) Encrypt(body []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println(err)
	}

	return gcm.Seal(nonce, nonce, body, nil), nil
}

func (c *Cipher) Decrypt(body []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.Key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(body) < nonceSize {
		return nil, err
	}

	nonce, ciphertext := body[:nonceSize], body[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func padKey(key []byte) []byte { // TODO Как то заполнить дополнительно
	newKey := make([]byte, 32)
	if len(key) > 32 {
		copy(newKey, key)
	} else if len(key) < 32 {
		times := math.Ceil(float64(32) / float64(len(key)))
		copy(newKey, bytes.Repeat(key, int(times)))
	} else {
		copy(newKey, key)
	}

	return newKey
}
