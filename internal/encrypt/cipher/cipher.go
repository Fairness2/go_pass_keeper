package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
)

type Cipher struct {
	Key []byte
}

func NewCipher(key []byte) *Cipher {
	return &Cipher{
		Key: key,
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
