package payloads

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCardWithComment_Encrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	expectErr := errors.New("expected error")
	tests := []struct {
		name           string
		card           CardWithComment
		getEncrypt     func() Encrypter
		wantErr        error
		expectedNumber []byte
		expectedDate   []byte
		expectedOwner  []byte
		expectedCVV    []byte
	}{
		{
			name: "valid_encryption",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte("01/25"),
					Owner:  []byte("Ivan Ivanow"),
					CVV:    []byte("123"),
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				second := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				third := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted3"), nil).Times(1).After(second)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted4"), nil).Times(1).After(third)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte("encrypted2"),
			expectedOwner:  []byte("encrypted3"),
			expectedCVV:    []byte("encrypted4"),
		},
		{
			name: "only_number",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_date",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte("123"),
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte("encrypted2"),
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_owner",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte("123"),
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte("encrypted2"),
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_cvv",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte("123"),
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte("encrypted2"),
		},

		{
			name: "error_number",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, expectErr).Times(1)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_error_date",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte("123"),
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, expectErr).Times(1).After(first)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte("encrypted2"),
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_error_owner",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte("123"),
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, expectErr).Times(1).After(first)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte("encrypted2"),
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_error_cvv",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte("123"),
				},
				Comment: "comment",
			},
			getEncrypt: func() Encrypter {
				en := NewMockEncrypter(ctr)
				first := en.EXPECT().Encrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Encrypt(gomock.Any()).Return(nil, expectErr).Times(1).After(first)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte("encrypted2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Encrypt(tt.getEncrypt())
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.card.Number, tt.expectedNumber)
				assert.Equal(t, tt.card.Date, tt.expectedDate)
				assert.Equal(t, tt.card.Owner, tt.expectedOwner)
				assert.Equal(t, tt.card.CVV, tt.expectedCVV)
			}
		})
	}
}

func TestCardWithComment_Decrypt(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	expectErr := errors.New("expected error")
	tests := []struct {
		name           string
		card           CardWithComment
		getEncrypt     func() Decrypter
		wantErr        error
		expectedNumber []byte
		expectedDate   []byte
		expectedOwner  []byte
		expectedCVV    []byte
	}{
		{
			name: "valid_encryption",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte("01/25"),
					Owner:  []byte("Ivan Ivanow"),
					CVV:    []byte("123"),
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				first := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				second := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				third := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted3"), nil).Times(1).After(second)
				en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted4"), nil).Times(1).After(third)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte("encrypted2"),
			expectedOwner:  []byte("encrypted3"),
			expectedCVV:    []byte("encrypted4"),
		},
		{
			name: "only_number",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_date",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte("123"),
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				first := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte("encrypted2"),
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_owner",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte("123"),
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				first := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte("encrypted2"),
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_cvv",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte("123"),
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				first := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted2"), nil).Times(1).After(first)
				return en
			},
			wantErr:        nil,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte("encrypted2"),
		},

		{
			name: "error_number",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				en.EXPECT().Decrypt(gomock.Any()).Return(nil, expectErr).Times(1)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_error_date",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte("123"),
					Owner:  []byte{},
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				first := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Decrypt(gomock.Any()).Return(nil, expectErr).Times(1).After(first)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte("encrypted2"),
			expectedOwner:  []byte{},
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_error_owner",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte("123"),
					CVV:    []byte{},
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				first := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Decrypt(gomock.Any()).Return(nil, expectErr).Times(1).After(first)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte("encrypted2"),
			expectedCVV:    []byte{},
		},
		{
			name: "number_and_error_cvv",
			card: CardWithComment{
				Card: Card{
					Number: []byte("1234567890123456"),
					Date:   []byte{},
					Owner:  []byte{},
					CVV:    []byte("123"),
				},
				Comment: "comment",
			},
			getEncrypt: func() Decrypter {
				en := NewMockDecrypter(ctr)
				first := en.EXPECT().Decrypt(gomock.Any()).Return([]byte("encrypted1"), nil).Times(1)
				en.EXPECT().Decrypt(gomock.Any()).Return(nil, expectErr).Times(1).After(first)
				return en
			},
			wantErr:        expectErr,
			expectedNumber: []byte("encrypted1"),
			expectedDate:   []byte{},
			expectedOwner:  []byte{},
			expectedCVV:    []byte("encrypted2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.card.Decrypt(tt.getEncrypt())
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.card.Number, tt.expectedNumber)
				assert.Equal(t, tt.card.Date, tt.expectedDate)
				assert.Equal(t, tt.card.Owner, tt.expectedOwner)
				assert.Equal(t, tt.card.CVV, tt.expectedCVV)
			}
		})
	}
}
