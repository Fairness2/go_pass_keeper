package user

import (
	"passkeeper/internal/client/config"
	"passkeeper/internal/encrypt/cipher"
)

// User представляет собой объект с идентификатором и паролем, обычно используемый для аутентификации и идентификации.
type User struct {
	ID     int64
	Cipher *cipher.Cipher
}

// CurrentUser содержит информацию о текущем аутентифицированном пользователе, включая идентификатор и пароль.
var CurrentUser *User

// SetUser инициализирует CurrentUser с предоставленным идентификатором и шифратором, используя заданный пароль и предопределенные конфигурации.
func SetUser(id int64, password string) {
	CurrentUser = &User{
		ID: id,
		Cipher: cipher.NewCipher(cipher.Config{
			Key:        []byte(password),
			Salt:       config.PassKeySalt,
			Iterations: config.PassKeyIterations,
			Length:     config.PassKeyLength,
		}),
	}
}
