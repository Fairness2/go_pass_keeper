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
