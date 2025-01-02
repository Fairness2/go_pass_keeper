package payloads

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	expErr := errors.New("error")
	tests := []struct {
		name         string
		initialText  []byte
		getEncrypt   func() Encrypter
		expectedText []byte
		expectError  bool
	}{
		{
			name:        "valid_encryption",
			initialText: []byte("test"),
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("test"), nil).Times(1)
				return en
			},
			expectedText: []byte("test"),
			expectError:  false,
		},
		{
			name:        "error_encryption",
			initialText: []byte(""),
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, expErr).Times(1)
				return en
			},
			expectedText: nil,
			expectError:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item := &TextWithComment{
				Text: Text{
					TextData: tc.initialText,
				},
			}

			err := item.Encrypt(tc.getEncrypt())
			if err != nil {
				assert.ErrorIs(t, expErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedText, item.Text.TextData)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	expErr := errors.New("error")
	tests := []struct {
		name          string
		encryptedText []byte
		getDecrypt    func() Decrypter
		expectedText  []byte
		expectError   bool
	}{
		{
			name:          "valid_decryption",
			encryptedText: []byte("encrypted"),
			getDecrypt: func() Decrypter {
				de := NewMockDecrypter(ctr)
				de.EXPECT().Decrypt(gomock.Any()).Return([]byte("decrypted"), nil).Times(1)
				return de
			},
			expectedText: []byte("decrypted"),
			expectError:  false,
		},
		{
			name:          "error_decryption",
			encryptedText: []byte("invalid"),
			getDecrypt: func() Decrypter {
				de := NewMockDecrypter(ctr)
				de.EXPECT().Decrypt(gomock.Any()).Return(nil, expErr).Times(1)
				return de
			},
			expectedText: nil,
			expectError:  true,
		},
		{
			name:          "empty_input",
			encryptedText: []byte{},
			getDecrypt: func() Decrypter {
				de := NewMockDecrypter(ctr)
				de.EXPECT().Decrypt(gomock.Any()).Return([]byte{}, nil).Times(1)
				return de
			},
			expectedText: []byte{},
			expectError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item := &TextWithComment{
				Text: Text{
					TextData: tc.encryptedText,
				},
			}

			err := item.Decrypt(tc.getDecrypt())
			if tc.expectError {
				assert.ErrorIs(t, expErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedText, item.Text.TextData)
			}
		})
	}
}
