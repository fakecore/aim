package errors

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorExitCodes(t *testing.T) {
	tests := []struct {
		category string
		want     int
	}{
		{"CFG", 2},
		{"ACC", 3},
		{"VEN", 4},
		{"TOO", 5},
		{"EXE", 6},
		{"", 1}, // default
	}

	for _, tt := range tests {
		e := &Error{Category: tt.category}
		assert.Equal(t, tt.want, e.ExitCode())
	}
}

func TestWrap(t *testing.T) {
	err := Wrap(ErrAccountNotFound, "deepseek")
	assert.Equal(t, "AIM-ACC-001", err.Code)
	assert.Contains(t, err.Message, "deepseek")
}
