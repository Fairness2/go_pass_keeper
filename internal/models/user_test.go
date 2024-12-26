package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCheckPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		userHash string
		input    string
		expected bool
		wantErr  bool
	}{
		{
			name:     "valid_matching_hashes",
			userHash: "5d41402abc4b2a76b9719d911017c592", // hashed value of 'hello'
			input:    "5d41402abc4b2a76b9719d911017c592", // same as userHash
			expected: true,
			wantErr:  false,
		},
		{
			name:     "non_matching_hashes",
			userHash: "5d41402abc4b2a76b9719d911017c592", // hashed value of 'hello'
			input:    "6d41402abc4b2a76b9719d911017c593", // different hash
			expected: false,
			wantErr:  false,
		},
		{
			name:     "invalid_input_hash",
			userHash: "5d41402abc4b2a76b9719d911017c592", // hashed value of 'hello'
			input:    "invalidhash",                      // invalid hex string
			expected: false,
			wantErr:  true,
		},
		{
			name:     "invalid_user_hash",
			userHash: "invaliduserhash",                  // invalid hex string
			input:    "5d41402abc4b2a76b9719d911017c592", // valid hex string
			expected: false,
			wantErr:  true,
		},
		{
			name:     "both_hashes_invalid",
			userHash: "invaliduserhash",
			input:    "invalidinputhash",
			expected: false,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				PasswordHash: tt.userHash,
			}
			result, err := user.CheckPasswordHash(tt.input)
			if tt.wantErr {
				assert.Error(t, err, "expected error state: got %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "unexpected error state: got %v, wantErr %v", err, tt.wantErr)
			assert.Equal(t, tt.expected, result, "expected result: got %v, want %v", result, tt.expected)
		})
	}
}

func TestGeneratePasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		password string
		hashKey  string
		wantErr  bool
	}{
		{
			name:     "valid_password_and_hash_key",
			password: "password123",
			hashKey:  "someHashKey",
			wantErr:  false,
		},
		{
			name:     "empty_password",
			password: "",
			hashKey:  "someHashKey",
			wantErr:  true,
		},
		{
			name:     "empty_hash_key",
			password: "password123",
			hashKey:  "",
			wantErr:  true,
		},
		{
			name:     "special_characters_in_password",
			password: "!@#$%^&*()_+=-",
			hashKey:  "specialKey",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := User{
				Password: tt.password,
			}
			err := user.GeneratePasswordHash(tt.hashKey)
			if tt.wantErr {
				assert.Error(t, err, "expected error state: got %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.NoError(t, err, "unexpected error state: got %v, wantErr %v", err, tt.wantErr)
			assert.NotEmpty(t, user.PasswordHash, "expected non-empty hash")
		})
	}
}
