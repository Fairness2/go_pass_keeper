package payloads

import "passkeeper/internal/encrypt/cipher"

// UpdateFile представляет собой полезную нагрузку запроса для обновления информации о файле, такой как его имя или связанный комментарий.
type UpdateFile struct {
	ID      int64  `json:"id,omitempty"`
	Name    []byte `json:"name"`
	Comment string `json:"comment"`
}

// FileWithComment представляет файл со связанным комментарием.
// Он содержит поля для идентификатора файла, его имени в виде байтового фрагмента и текстового комментария.
type FileWithComment struct {
	ID      int64  `json:"id"`
	Name    []byte `json:"name"`
	Comment string `json:"comment"`
}

func (item *FileWithComment) Encrypt(ch *cipher.Cipher) error {
	eName, err := ch.Encrypt(item.Name)
	if err != nil {
		return err
	}
	item.Name = eName
	return nil
}

func (item *FileWithComment) Decrypt(ch *cipher.Cipher) error {
	dName, err := ch.Decrypt(item.Name)
	if err != nil {
		return err
	}
	item.Name = dName
	return nil
}
