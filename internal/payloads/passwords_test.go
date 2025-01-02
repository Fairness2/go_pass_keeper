package payloads

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPasswordWithComment_Encrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	usernameErr := errors.New("username error")
	passwordErr := errors.New("password error")
	tests := []struct {
		name             string
		item             PasswordWithComment
		getEncrypt       func() Encrypter
		expectErr        error
		expectedUsername []byte
		expectedPassword []byte
	}{
		{
			name: "successful_encryption",
			item: PasswordWithComment{
				Password: Password{Username: []byte("username"), Password: []byte("password")},
				Comment:  "test comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("Username"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("Password"), nil).Times(1).After(first)
				return en
			},
			expectedUsername: []byte("Username"),
			expectedPassword: []byte("Password"),
			expectErr:        nil,
		},
		{
			name: "fail_encrypting_username",
			item: PasswordWithComment{
				Password: Password{Username: []byte("username"), Password: []byte("password")},
				Comment:  "test comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, usernameErr).Times(1)
				return en
			},
			expectedUsername: nil,
			expectedPassword: nil,
			expectErr:        usernameErr,
		},
		{
			name: "successful_encryption",
			item: PasswordWithComment{
				Password: Password{Username: []byte("username"), Password: []byte("password")},
				Comment:  "test comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("test"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, passwordErr).Times(1).After(first)
				return en
			},
			expectedUsername: []byte("Username"),
			expectedPassword: []byte("Password"),
			expectErr:        passwordErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Encrypt(tt.getEncrypt())
			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr)

			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUsername, tt.item.Password.Username)
				assert.Equal(t, tt.expectedPassword, tt.item.Password.Password)
			}
		})
	}
}

func TestPasswordWithComment_Decrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	usernameErr := errors.New("username decryption error")
	passwordErr := errors.New("password decryption error")
	tests := []struct {
		name             string
		item             PasswordWithComment
		getDecrypt       func() Decrypter
		expectErr        error
		expectedUsername []byte
		expectedPassword []byte
	}{
		{
			name: "successful_decryption",
			item: PasswordWithComment{
				Password: Password{Username: []byte("EncryptedUsername"), Password: []byte("EncryptedPassword")},
				Comment:  "test comment",
			},
			getDecrypt: func() Decrypter {
				dc := NewMockDecrypter(ctr)
				first := dc.EXPECT().Decrypt(gomock.Any()).Return([]byte("DecryptedUsername"), nil).Times(1)
				dc.EXPECT().Decrypt(gomock.Any()).Return([]byte("DecryptedPassword"), nil).Times(1).After(first)
				return dc
			},
			expectedUsername: []byte("DecryptedUsername"),
			expectedPassword: []byte("DecryptedPassword"),
			expectErr:        nil,
		},
		{
			name: "fail_decrypting_username",
			item: PasswordWithComment{
				Password: Password{Username: []byte("EncryptedUsername"), Password: []byte("EncryptedPassword")},
				Comment:  "test comment",
			},
			getDecrypt: func() Decrypter {
				dc := NewMockDecrypter(ctr)
				dc.EXPECT().Decrypt(gomock.Any()).Return(nil, usernameErr).Times(1)
				return dc
			},
			expectedUsername: nil,
			expectedPassword: nil,
			expectErr:        usernameErr,
		},
		{
			name: "fail_decrypting_password",
			item: PasswordWithComment{
				Password: Password{Username: []byte("EncryptedUsername"), Password: []byte("EncryptedPassword")},
				Comment:  "test comment",
			},
			getDecrypt: func() Decrypter {
				dc := NewMockDecrypter(ctr)
				first := dc.EXPECT().Decrypt(gomock.Any()).Return([]byte("DecryptedUsername"), nil).Times(1)
				dc.EXPECT().Decrypt(gomock.Any()).Return(nil, passwordErr).Times(1).After(first)
				return dc
			},
			expectedUsername: []byte("DecryptedUsername"),
			expectedPassword: nil,
			expectErr:        passwordErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Decrypt(tt.getDecrypt())
			if tt.expectErr != nil {
				assert.ErrorIs(t, err, tt.expectErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUsername, tt.item.Password.Username)
				assert.Equal(t, tt.expectedPassword, tt.item.Password.Password)
			}
		})
	}
}
