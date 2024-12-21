package payloads

import "passkeeper/internal/encrypt/cipher"

// SaveText представляет собой структуру для получения текстовых данных с необязательным комментарием и идентификатором.
type SaveText struct {
	ID       int64  `json:"id,omitempty"`
	TextData []byte `json:"text_data"`
	Comment  string `json:"comment"`
}

// Text представляет текстовый объект с идентификатором и двоичными текстовыми данными.
type Text struct {
	ID       int64  `json:"id"`
	TextData []byte `json:"text_data"`
}

// TextWithComment — это структура, которая объединяет текстовый объект с дополнительным комментарием, представленным в виде строки.
type TextWithComment struct {
	Text
	Comment string `json:"comment"`
}

func (item *TextWithComment) Encrypt(ch *cipher.Cipher) error {
	eText, err := ch.Encrypt(item.TextData)
	if err != nil {
		return err
	}
	item.TextData = eText
	return nil
}

func (item *TextWithComment) Decrypt(ch *cipher.Cipher) error {
	eText, err := ch.Decrypt(item.TextData)
	if err != nil {
		return err
	}
	item.TextData = eText
	return nil
}
