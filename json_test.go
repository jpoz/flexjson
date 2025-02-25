package flexjson

import (
	"reflect"
	"testing"
)

func TestParsePartialJSONObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
		wantErr  bool
	}{
		// Complete JSON objects
		{
			name:     "Empty object",
			input:    "{}",
			expected: map[string]any{},
			wantErr:  false,
		},
		{
			name:     "Complete simple key-value",
			input:    `{"key": 123}`,
			expected: map[string]any{"key": int64(123)},
			wantErr:  false,
		},
		{
			name:  "Complete multiple key-values",
			input: `{"key1": 123, "key2": "value", "key3": true}`,
			expected: map[string]any{
				"key1": int64(123),
				"key2": "value",
				"key3": true,
			},
			wantErr: false,
		},

		// Partial JSON cases
		{
			name:     "Partial simple key-value without closing brace",
			input:    `{"key": 123`,
			expected: map[string]any{"key": int64(123)},
			wantErr:  false,
		},
		{
			name:     "Partial key-value with incomplete second key",
			input:    `{"key1": 1234, "key2":`,
			expected: map[string]any{"key1": int64(1234), "key2": nil},
			wantErr:  false,
		},
		{
			name:     "Partial with trailing comma",
			input:    `{"key1": "value", "key2": false,`,
			expected: map[string]any{"key1": "value", "key2": false},
			wantErr:  false,
		},
		{
			name:     "Partial with incomplete key (no colon)",
			input:    `{"key1": true, "key2`,
			expected: map[string]any{"key1": true, "key2": nil},
			wantErr:  false,
		},

		// Nested objects and arrays
		{
			name:     "Complete nested object",
			input:    `{"key1": {"nested": 42}}`,
			expected: map[string]any{"key1": map[string]any{"nested": int64(42)}},
			wantErr:  false,
		},
		{
			name:     "Partial nested object",
			input:    `{"key1": {"nested": 42`,
			expected: map[string]any{"key1": map[string]any{"nested": int64(42)}},
			wantErr:  false,
		},
		{
			name:     "Complete with array",
			input:    `{"key1": [1, 2, 3]}`,
			expected: map[string]any{"key1": []interface{}{int64(1), int64(2), int64(3)}},
			wantErr:  false,
		},
		{
			name:     "Partial array",
			input:    `{"key1": [1, 2,`,
			expected: map[string]any{"key1": []interface{}{int64(1), int64(2)}},
			wantErr:  false,
		},

		// Error cases
		{
			name:     "Not an object",
			input:    `[1, 2, 3]`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Empty string",
			input:    ``,
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePartialJSONObject(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePartialJSONObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePartialJSONObject() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Test the exact examples from the requirements
func TestRequirementExamples(t *testing.T) {
	// Example 1: {"key": 123 should parse into map[string]any{"key": 123}
	example1, err := ParsePartialJSONObject(`{"key": 123`)
	if err != nil {
		t.Errorf("Failed on example 1: %v", err)
	}
	expected1 := map[string]any{"key": int64(123)}
	if !reflect.DeepEqual(example1, expected1) {
		t.Errorf("Example 1 result = %v, want %v", example1, expected1)
	}

	// Example 2: {"key": 1234, "key2": should parse into map[string]any{"key": 1234, "key2": nil}
	example2, err := ParsePartialJSONObject(`{"key": 1234, "key2":`)
	if err != nil {
		t.Errorf("Failed on example 2: %v", err)
	}
	expected2 := map[string]any{"key": int64(1234), "key2": nil}
	if !reflect.DeepEqual(example2, expected2) {
		t.Errorf("Example 2 result = %v, want %v", example2, expected2)
	}
}

func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]any
		wantErr  bool
	}{
		{
			name:  "Nested incomplete objects",
			input: `{"obj1": {"key1": 1, "obj2": {"key2":`,
			expected: map[string]any{
				"obj1": map[string]any{
					"key1": int64(1),
					"obj2": map[string]any{
						"key2": nil,
					},
				},
			},
			wantErr: false,
		},
		{
			name:  "Mixed types",
			input: `{"str": "hello", "num": 42, "bool": true, "null": null, "arr": [1,`,
			expected: map[string]any{
				"str":  "hello",
				"num":  int64(42),
				"bool": true,
				"null": nil,
				"arr":  []interface{}{int64(1)},
			},
			wantErr: false,
		},
		{
			name:  "Unicode characters",
			input: `{"unicode": "😀🌍🚀", "key2":`,
			expected: map[string]any{
				"unicode": "😀🌍🚀",
				"key2":    nil,
			},
			wantErr: false,
		},
		{
			name:  "Deeply nested structure",
			input: `{"l1": {"l2": {"l3": {"l4": {"l5":`,
			expected: map[string]any{
				"l1": map[string]any{
					"l2": map[string]any{
						"l3": map[string]any{
							"l4": map[string]any{
								"l5": nil,
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePartialJSONObject(tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParsePartialJSONObject() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("ParsePartialJSONObject() = %v, want %v", result, tt.expected)
			}
		})
	}
}
