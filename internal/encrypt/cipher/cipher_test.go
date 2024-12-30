package cipher

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPadKey(t *testing.T) {
	tests := []struct {
		name       string
		key        []byte
		salt       []byte
		iterations int
		length     int
		expected   []byte
	}{
		{
			name:       "key_length_less_than_32_bytes",
			key:        []byte("short_key"),
			salt:       []byte("salt"),
			iterations: 100000,
			length:     32,
			expected:   []byte{0x81, 0x77, 0xbe, 0xeb, 0x88, 0x8, 0xf2, 0x9d, 0x73, 0x3, 0xc8, 0x22, 0xdd, 0xac, 0xc6, 0xf3, 0xc8, 0xcd, 0x88, 0x8d, 0x22, 0x86, 0x0, 0x6e, 0x0, 0x94, 0xab, 0xd3, 0x36, 0x9, 0xe6, 0x49},
		},
		{
			name:       "empty_key",
			key:        []byte{},
			salt:       []byte("salt"),
			iterations: 100000,
			length:     32,
			expected:   []byte{0x37, 0x8a, 0x8b, 0xda, 0x97, 0x6a, 0xde, 0xa, 0x44, 0x97, 0xd2, 0xa0, 0x4b, 0xd1, 0xce, 0x35, 0x2b, 0xc6, 0x6c, 0x38, 0x99, 0x47, 0x62, 0x4a, 0xe5, 0xfa, 0xb8, 0xc6, 0x20, 0x74, 0xa3, 0x71},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := padKey(tc.key, tc.salt, tc.iterations, tc.length)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestNewCipher(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		salt        []byte
		iterations  int
		length      int
		expectedKey []byte
	}{
		{
			name:        "valid_short_key",
			key:         []byte("short_key"),
			salt:        []byte("salt"),
			iterations:  100000,
			length:      32,
			expectedKey: []byte{0x81, 0x77, 0xbe, 0xeb, 0x88, 0x8, 0xf2, 0x9d, 0x73, 0x3, 0xc8, 0x22, 0xdd, 0xac, 0xc6, 0xf3, 0xc8, 0xcd, 0x88, 0x8d, 0x22, 0x86, 0x0, 0x6e, 0x0, 0x94, 0xab, 0xd3, 0x36, 0x9, 0xe6, 0x49},
		},
		{
			name:        "empty_key_input",
			key:         []byte(""),
			salt:        []byte("salt"),
			iterations:  100000,
			length:      32,
			expectedKey: []byte{0x37, 0x8a, 0x8b, 0xda, 0x97, 0x6a, 0xde, 0xa, 0x44, 0x97, 0xd2, 0xa0, 0x4b, 0xd1, 0xce, 0x35, 0x2b, 0xc6, 0x6c, 0x38, 0x99, 0x47, 0x62, 0x4a, 0xe5, 0xfa, 0xb8, 0xc6, 0x20, 0x74, 0xa3, 0x71},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cipher := NewCipher(Config{
				Key:        tc.key,
				Salt:       tc.salt,
				Iterations: tc.iterations,
				Length:     tc.length,
			})
			assert.Equal(t, tc.expectedKey, cipher.key)
		})
	}
}

func TestEncrypt(t *testing.T) {
	tests := []struct {
		name          string
		key           []byte
		plaintext     []byte
		expectedError bool
	}{
		{
			name:          "encrypt_non_empty_plaintext_with_valid_key",
			key:           []byte("this_is_a_32_byte_key_0123456789"),
			plaintext:     []byte("This is a plaintext message."),
			expectedError: false,
		},
		{
			name:          "long_message_with_valid_key",
			key:           []byte("this_is_a_32_byte_key_0123456789"),
			plaintext:     []byte("This is a long message.This is a long message.This is a long message.This is a long message.This is a long message.This is a long message.This is a long message.This is a long message.This is a long message.This is a long message."),
			expectedError: false,
		},
		{
			name:          "incorrect_key_length",
			key:           []byte("this_is_a_32_byte_key9"),
			plaintext:     []byte("This is a plaintext message."),
			expectedError: true,
		},
		{
			name:          "empty_plaintext",
			key:           []byte("this_is_a_32_byte_key_0123456789"),
			plaintext:     []byte(""),
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cipher := Cipher{key: tc.key}

			encrypted, err := cipher.Encrypt(tc.plaintext)

			if tc.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, encrypted)
			}
		})
	}
}

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name          string
		key           []byte
		encryptedData []byte
		expected      []byte
		expectedError bool
	}{
		{
			name: "valid_message_decryption",
			key:  []byte("this_is_a_32_byte_key_0123456789"),
			encryptedData: func() []byte {
				cipher := Cipher{key: []byte("this_is_a_32_byte_key_0123456789")}
				data, _ := cipher.Encrypt([]byte("This is a plaintext message."))
				return data
			}(),
			expected:      []byte("This is a plaintext message."),
			expectedError: false,
		},
		{
			name: "wrong_key",
			key:  []byte("this_is_a_32_byte_key_0123456780"),
			encryptedData: func() []byte {
				cipher := Cipher{key: []byte("this_is_a_32_byte_key_0123456789")}
				data, _ := cipher.Encrypt([]byte("This is a plaintext message."))
				return data
			}(),
			expected:      nil,
			expectedError: true,
		},
		{
			name: "tampered_ciphertext",
			key:  []byte("this_is_a_32_byte_key_0123456789"),
			encryptedData: func() []byte {
				cipher := Cipher{key: []byte("this_is_a_32_byte_key_0123456789")}
				data, _ := cipher.Encrypt([]byte("This is a plaintext message."))
				data[len(data)-1] ^= 0xFF // Tamper the ciphertext
				return data
			}(),
			expected:      nil,
			expectedError: true,
		},
		{
			name:          "empty_input",
			key:           []byte("this_is_a_32_byte_key_0123456789"),
			encryptedData: []byte{},
			expected:      nil,
			expectedError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cipher := Cipher{key: tc.key}

			decrypted, err := cipher.Decrypt(tc.encryptedData)

			if tc.expectedError {
				assert.Error(t, err)
				assert.Nil(t, decrypted)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, decrypted)
			}
		})
	}
}
