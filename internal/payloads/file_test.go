package payloads

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFileWithComment_Encrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	expectErr := errors.New("expected error")
	tests := []struct {
		name         string
		inputName    []byte
		expectedName []byte
		expectError  error
		getEncrypt   func() Encrypter
	}{
		{
			name:         "successful encryption",
			inputName:    []byte("testfile"),
			expectedName: []byte("encryptedtestfile"),
			expectError:  nil,
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encryptedtestfile"), nil).Times(1)
				return en
			},
		},
		{
			name:         "encryption error",
			inputName:    []byte("testfile"),
			expectedName: []byte("testfile"),
			expectError:  expectErr,
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, expectErr).Times(1)
				return en
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &FileWithComment{
				ID:      1,
				Name:    tt.inputName,
				Comment: "example comment",
			}

			err := item.Encrypt(tt.getEncrypt())
			if err != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedName, item.Name)
			}
		})
	}
}

func TestFileWithComment_Decrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	expectErr := errors.New("expected error")
	tests := []struct {
		name         string
		inputName    []byte
		expectedName []byte
		expectError  error
		getDecrypt   func() Decrypter
	}{
		{
			name:         "successful decryption",
			inputName:    []byte("encryptedtestfile"),
			expectedName: []byte("testfile"),
			expectError:  nil,
			getDecrypt: func() Decrypter {
				dc := NewMockDecrypter(ctr)
				dc.EXPECT().Decrypt(gomock.Any()).Return([]byte("testfile"), nil).Times(1)
				return dc
			},
		},
		{
			name:         "decryption error",
			inputName:    []byte("encryptedtestfile"),
			expectedName: []byte("encryptedtestfile"),
			expectError:  expectErr,
			getDecrypt: func() Decrypter {
				dc := NewMockDecrypter(ctr)
				dc.EXPECT().Decrypt(gomock.Any()).Return(nil, expectErr).Times(1)
				return dc
			},
		},
		{
			name:         "empty input",
			inputName:    []byte{},
			expectedName: []byte{},
			expectError:  nil,
			getDecrypt: func() Decrypter {
				dc := NewMockDecrypter(ctr)
				dc.EXPECT().Decrypt(gomock.Any()).Return([]byte{}, nil).Times(1)
				return dc
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := &FileWithComment{
				ID:      1,
				Name:    tt.inputName,
				Comment: "example comment",
			}

			err := item.Decrypt(tt.getDecrypt())
			if err != nil {
				assert.Equal(t, tt.expectError, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, tt.expectedName, item.Name)
			}
		})
	}
}
