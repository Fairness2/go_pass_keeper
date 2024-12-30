package payloads

// Register пэйлоад для регистрации пользователя
type Register struct {
	Login    string `json:"login" valid:"required,type(string),minstringlength(3)"`
	Password string `json:"password" valid:"required,type(string),minstringlength(6)"`
}

// Login пэйлоад для авторизации пользователя
type Login struct {
	Login    string `json:"login" valid:"required,type(string)"`
	Password string `json:"password" valid:"required,type(string)"`
}

// Authorization ответ с токеном авторизации
type Authorization struct {
	Token   string `json:"token"`
	Refresh string `json:"refresh"`
}
