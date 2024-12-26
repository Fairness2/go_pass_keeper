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

// Cipher — это структура, в которой хранится криптографический ключ для операций шифрования и дешифрования.
type Cipher struct {
	key []byte
}

// NewCipher инициализирует новый экземпляр Cipher с помощью криптографического ключа, полученного из предоставленного ключа.
func NewCipher(key []byte) *Cipher {
	return &Cipher{
		key: padKey(key),
	}
}

// Encrypt шифрует данный фрагмент байта с использованием AES-GCM со случайно сгенерированным одноразовым номером и возвращает зашифрованные данные.
func (c *Cipher) Encrypt(body []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
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

// Decrypt расшифровывает предоставленный фрагмент байта с помощью AES-GCM и возвращает открытый текст или ошибку, если расшифровка не удалась.
func (c *Cipher) Decrypt(body []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
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

// padKey настраивает предоставленный ключ до фиксированной длины в 32 байта, усекая или дополняя его повторяющимися байтами.
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
