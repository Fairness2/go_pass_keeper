package cipher

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPadKey(t *testing.T) {
	tests := []struct {
		name     string
		key      []byte
		expected []byte
	}{
		{
			name:     "key_length_less_than_32_bytes",
			key:      []byte("short_key"),
			expected: append(bytes.Repeat([]byte("short_key"), 4)[:32]),
		},
		{
			name:     "key_length_exactly_32_bytes",
			key:      []byte("this_is_a_32_byte_key_0123456789"),
			expected: []byte("this_is_a_32_byte_key_0123456789"),
		},
		{
			name:     "key_length_greater_than_32_bytes",
			key:      []byte("this_is_a_key_that_is_longer_than_32_bytes"),
			expected: []byte("this_is_a_key_that_is_longer_tha"),
		},
		{
			name:     "key_of_1_byte",
			key:      []byte("A"),
			expected: append(bytes.Repeat([]byte("A"), 32)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := padKey(tc.key)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func TestNewCipher(t *testing.T) {
	tests := []struct {
		name        string
		key         []byte
		expectedKey []byte
	}{
		{
			name:        "valid_short_key",
			key:         []byte("shortkey"),
			expectedKey: append(bytes.Repeat([]byte("shortkey"), 4)[:32]),
		},
		{
			name:        "valid_32_byte_key",
			key:         []byte("this_is_a_32_byte_key_0123456789"),
			expectedKey: []byte("this_is_a_32_byte_key_0123456789"),
		},
		{
			name:        "key_longer_than_32_bytes",
			key:         []byte("this_is_a_key_that_is_longer_than_32_bytes_12345"),
			expectedKey: []byte("this_is_a_key_that_is_longer_tha"),
		},
		{
			name:        "empty_key_input",
			key:         []byte(""),
			expectedKey: bytes.Repeat([]byte{'\x00'}, 32),
		},
		{
			name:        "single_byte_key",
			key:         []byte("A"),
			expectedKey: bytes.Repeat([]byte("A"), 32),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cipher := NewCipher(tc.key)
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
