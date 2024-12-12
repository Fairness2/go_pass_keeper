package encrypt

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
)

// Encrypter Класс для дешифрования данных
type Encrypter struct {
	publicKey *rsa.PublicKey
}

// NewEncrypter Создаёт новый Encrypter с переданным ключом
func NewEncrypter(publicKey *rsa.PublicKey) *Encrypter {
	return &Encrypter{publicKey: publicKey}
}

// Encrypt шифрует данное тело, используя шифрование RSA с открытым ключом из структуры пула.
// Возвращает зашифрованное тело или ошибку, если шифрование не удалось.
func (e *Encrypter) Encrypt(body []byte) ([]byte, error) {
	if e.publicKey == nil {
		return body, ErrorEmptyKey
	}
	label := []byte("")
	hash := sha256.New()
	blockSize := e.publicKey.Size() - 2*hash.Size() - 2
	blocks := splitMessage(body, blockSize)
	encryptedBlocks := make([][]byte, len(blocks))
	for i, block := range blocks {
		newBlock, err := rsa.EncryptOAEP(hash, rand.Reader, e.publicKey, block, label)
		if err != nil {
			return body, err
		}
		encryptedBlocks[i] = newBlock
	}
	newBody := bytes.Join(encryptedBlocks, nil)
	return newBody, nil
}
