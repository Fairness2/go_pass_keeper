package payloads

// SavePassword представляет структуру для сохранения пароля, включая домен, имя пользователя, пароль и необязательный комментарий.
type SavePassword struct {
	ID       int64  `json:"id,omitempty"`
	Domen    string `json:"domen"`
	Username []byte `json:"username"`
	Password []byte `json:"password"`
	Comment  string `json:"comment"`
}

// Password представляет структуру записи пароля с полями идентификатора, домена, имени пользователя и пароля.
type Password struct {
	ID       int64  `json:"id"`
	Domen    string `json:"domen"`
	Username []byte `json:"username"`
	Password []byte `json:"password"`
}

// PasswordWithComment представляет собой структуру, которая расширяет Password дополнительным полем комментария.
type PasswordWithComment struct {
	Password
	Comment string `json:"comment"`
}

// Encrypt шифрует поля имени пользователя и пароля объекта PasswordWithComment, используя предоставленный экземпляр Encrypter.
// Возвращает ошибку, если шифрование любого поля не удалось.
func (item *PasswordWithComment) Encrypt(ch Encrypter) error {
	eUsername, err := ch.Encrypt(item.Username)
	if err != nil {
		return err
	}
	item.Username = eUsername
	ePass, err := ch.Encrypt(item.Password.Password)
	if err != nil {
		return err
	}
	item.Password.Password = ePass
	return nil
}

// Decrypt расшифровывает поля имени пользователя и пароля объекта PasswordWithComment с помощью предоставленного расшифровщика.
// Возвращает ошибку, если расшифровка любого поля не удалась.
func (item *PasswordWithComment) Decrypt(ch Decrypter) error {
	dUsername, err := ch.Decrypt(item.Username)
	if err != nil {
		return err
	}
	item.Username = dUsername
	dPass, err := ch.Decrypt(item.Password.Password)
	if err != nil {
		return err
	}
	item.Password.Password = dPass
	return nil
}
