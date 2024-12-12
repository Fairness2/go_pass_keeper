package payloads

type SavePassword struct {
	ID       int64  `json:"id,omitempty"`
	Domen    string `json:"domen"`
	Username string `json:"username"`
	Password string `json:"password"`
	Comment  string `json:"comment"`
}

type Password struct {
	ID       int64  `json:"id"`
	Domen    string `json:"domen"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type PasswordWithComment struct {
	Password
	Comment string `json:"comment"`
}
