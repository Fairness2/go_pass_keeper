package encrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"errors"
	"math"
)

// ErrorEmptyKey Ошибка, что был передан пустой ключ для шифрования или дешифрования
var ErrorEmptyKey = errors.New("empty key")

// Decrypter Класс для дешифрования данных
type Decrypter struct {
	privateKey *rsa.PrivateKey
}

// NewDecrypter Создаёт новый Decrypter с переданным ключом
func NewDecrypter(privateKey *rsa.PrivateKey) *Decrypter {
	return &Decrypter{privateKey: privateKey}
}

// decrypt Дешифрование переданого тела
func (d Decrypter) Decrypt(message []byte) ([]byte, error) {
	if d.privateKey == nil {
		return nil, ErrorEmptyKey
	}
	label := []byte("")
	hash := sha256.New()
	encryptedBlocks := splitMessage(message, d.privateKey.Size())
	blocks := make([][]byte, len(encryptedBlocks))
	for i, block := range encryptedBlocks {
		newBlock, err := rsa.DecryptOAEP(hash, rand.Reader, d.privateKey, block, label)
		if err != nil {
			return nil, err
		}
		blocks[i] = newBlock
	}
	return bytes.Join(blocks, nil), nil
}

// Разделение текста на блоки нужного размера
func splitMessage(body []byte, blockSize int) [][]byte {
	var ln = math.Ceil(float64(len(body)) / float64(blockSize))
	blocks := make([][]byte, 0, int(ln))
	for i := 0; i < len(body); i += blockSize {
		end := i + blockSize
		if end > len(body) {
			end = len(body)
		}
		blocks = append(blocks, body[i:end])
	}
	return blocks
}
