package models

import (
	"time"
)

type PasswordContent struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	Domen     string    `db:"domen"`
	Username  []byte    `db:"username"`
	Password  []byte    `db:"password"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
	//EnPassword []byte    `db:"password"`
	//EnUsername []byte    `db:"username"`
}

/*func (p *PasswordContent) EncryptPrivateFields(encrypter *encrypt.Encrypter) error {
	if p.Password != "" {
		encr, err := encrypter.Encrypt([]byte(p.Password))
		if err != nil {
			return err
		}
		p.EnPassword = encr
	}
	if p.Username != "" {
		encr, err := encrypter.Encrypt([]byte(p.Username))
		if err != nil {
			return err
		}
		p.EnUsername = encr
	}
	return nil
}

func (p *PasswordContent) DecryptPrivateFields(decrypter *encrypt.Decrypter) error {
	if p.EnPassword != nil {
		dec, err := decrypter.Decrypt(p.EnPassword)
		if err != nil {
			return err
		}
		p.Password = string(dec)
	}
	if p.EnUsername != nil {
		dec, err := decrypter.Decrypt(p.EnUsername)
		if err != nil {
			return err
		}
		p.Username = string(dec)
	}
	return nil
}*/

type PasswordWithComment struct {
	PasswordContent
	Comment string `db:"comment"`
}
