package cipher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"golang.org/x/crypto/pbkdf2"
	"io"
)

type Config struct {
	Key        []byte
	Salt       []byte
	Iterations int
	Length     int
}

// Cipher — это структура, в которой хранится криптографический ключ для операций шифрования и дешифрования.
type Cipher struct {
	key []byte
}

// NewCipher инициализирует новый экземпляр Cipher с помощью криптографического ключа, полученного из предоставленного ключа.
func NewCipher(cnf Config) *Cipher {
	return &Cipher{
		key: padKey(cnf.Key, cnf.Salt, cnf.Iterations, cnf.Length),
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
func padKey(pass, salt []byte, iterations, length int) []byte { // TODO Как то заполнить дополнительно
	// Нам нужно при каждом хэшировании получать одинаковый  результат, чтобы не менялся ключ для шифрования разного контента
	return pbkdf2.Key(pass, salt, iterations, length, sha256.New)
}
