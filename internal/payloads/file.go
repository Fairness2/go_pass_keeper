package payloads

import _ "passkeeper/internal/validators"

// UpdateFile представляет собой полезную нагрузку запроса для обновления информации о файле, такой как его имя или связанный комментарий.
type UpdateFile struct {
	ID      string `json:"id,omitempty" valid:"required,uuidv4"`
	Name    []byte `json:"name" valid:"requireByteArray"`
	Comment string `json:"comment"`
}

// FileWithComment представляет файл со связанным комментарием.
// Он содержит поля для идентификатора файла, его имени в виде байтового фрагмента и текстового комментария.
type FileWithComment struct {
	ID      string `json:"id"`
	Name    []byte `json:"name"`
	Comment string `json:"comment"`
}

// Encrypt шифрует поле имени FileWithComment, используя предоставленный шифратор. Возвращает ошибку, если шифрование не удалось.
func (item *FileWithComment) Encrypt(ch Encrypter) error {
	eName, err := ch.Encrypt(item.Name)
	if err != nil {
		return err
	}
	item.Name = eName
	return nil
}

// Decrypt расшифровывает поле имени FileWithComment, используя предоставленный дешифратор. Возвращает ошибку, если расшифровка не удалась.
func (item *FileWithComment) Decrypt(ch Decrypter) error {
	dName, err := ch.Decrypt(item.Name)
	if err != nil {
		return err
	}
	item.Name = dName
	return nil
}
