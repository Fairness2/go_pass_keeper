package validators

import (
	"testing"
)

func TestArrayNotEmpty(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{name: "valid_non_empty_byte_slice", input: []byte{0x01, 0x02}, expected: true},
		{name: "empty_byte_slice", input: []byte{}, expected: false},
		{name: "nil_input", input: nil, expected: false},
		{name: "non_byte_slice_input", input: []int{1, 2, 3}, expected: false},
		{name: "unsupported_input_type_string", input: "test", expected: false},
		{name: "unsupported_input_type_int", input: 123, expected: false},
		{name: "nested_empty_byte_slice", input: [][]byte{{}}, expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := arrayNotEmpty(tt.input, nil)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
