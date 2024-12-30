package payloads

import _ "passkeeper/internal/validators"

// SaveText представляет собой структуру для сохранения текстовых данных с необязательным комментарием и идентификатором.
type SaveText struct {
	TextData []byte `json:"text_data" valid:"requireByteArray"`
	Comment  string `json:"comment" valid:"type(string)"`
}

// UpdateText представляет собой структуру для обновления текстовых данных с необязательным комментарием и идентификатором.
type UpdateText struct {
	ID       string `json:"id,omitempty" valid:"required,uuidv4"`
	TextData []byte `json:"text_data" valid:"requireByteArray"`
	Comment  string `json:"comment" valid:"type(string)"`
}

// Text представляет текстовый объект с идентификатором и двоичными текстовыми данными.
type Text struct {
	ID       string `json:"id"`
	TextData []byte `json:"text_data"`
}

// TextWithComment — это структура, которая объединяет текстовый объект с дополнительным комментарием, представленным в виде строки.
type TextWithComment struct {
	Text
	Comment string `json:"comment"`
}

// Encrypt шифрует поле TextData структуры TextWithComment, используя предоставленную реализацию Encrypter.
func (item *TextWithComment) Encrypt(ch Encrypter) error {
	eText, err := ch.Encrypt(item.TextData)
	if err != nil {
		return err
	}
	item.TextData = eText
	return nil
}

// Decrypt расшифровывает поле TextData структуры TextWithComment, используя предоставленную реализацию Decrypter.
func (item *TextWithComment) Decrypt(ch Decrypter) error {
	eText, err := ch.Decrypt(item.TextData)
	if err != nil {
		return err
	}
	item.TextData = eText
	return nil
}
