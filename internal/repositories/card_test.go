package repositories

import (
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCardRepository(t *testing.T) {
	ctr := gomock.NewController(t)
	defer ctr.Finish()
	tests := []struct {
		name string
		db   SQLExecutor
	}{
		{
			name: "valid_sqlexecutor",
			db:   NewMockSQLExecutor(ctr),
		},
		{
			name: "nil_sqlexecutor",
			db:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewCardRepository(tt.db)
			assert.NotNil(t, repo)
		})
	}
}
