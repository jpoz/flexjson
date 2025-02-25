package flexjson

import (
	"reflect"
	"testing"
)

func TestStreamingParser_SimpleObject(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	// Process a simple JSON object character by character
	json := `{"name":"John","age":30}`

	for _, char := range json {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Check the result
	expected := map[string]any{
		"name": "John",
		"age":  int64(30),
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Unexpected result. Got %v, expected %v", output, expected)
	}
}

func TestStreamingParser_NestedObject(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)
	sp.SetDebug(true)

	// Process a nested JSON object character by character
	json := `{"person":{"name":"John","age":30},"active":true}`

	for i, char := range json {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing %dth character '%c': %v", i, char, err)
		}
	}

	// Check the result
	expected := map[string]any{
		"person": map[string]any{
			"name": "John",
			"age":  int64(30),
		},
		"active": true,
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Unexpected result. Got %v, expected %v", output, expected)
	}
}

func TestStreamingParser_Array(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	// Process a JSON array character by character
	json := `{"numbers":[1,2,3],"names":["John","Jane"]}`

	for _, char := range json {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Check the result
	expected := map[string]any{
		"numbers": &[]interface{}{int64(1), int64(2), int64(3)},
		"names":   &[]interface{}{"John", "Jane"},
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Unexpected result. Got %v, expected %v", output, expected)
	}
}

func TestStreamingParser_ComplexTypes(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	// Process JSON with various types character by character
	json := `{"string":"hello","number":42,"float":3.14,"bool":true,"null":null}`

	for _, char := range json {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Check the result
	expected := map[string]any{
		"string": "hello",
		"number": int64(42),
		"float":  3.14,
		"bool":   true,
		"null":   nil,
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Unexpected result. Got %v, expected %v", output, expected)
	}
}

func TestStreamingParser_PartialJSON(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	// Process the first part of the JSON
	json1 := `{"name":"John"`

	for _, char := range json1 {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Check intermediate result
	expected1 := map[string]any{
		"name": "John",
	}

	if !reflect.DeepEqual(output, expected1) {
		t.Errorf("Unexpected intermediate result. Got %v, expected %v", output, expected1)
	}

	// Process the second part
	json2 := `,"age":30}`

	for _, char := range json2 {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Check final result
	expected2 := map[string]any{
		"name": "John",
		"age":  int64(30),
	}

	if !reflect.DeepEqual(output, expected2) {
		t.Errorf("Unexpected final result. Got %v, expected %v", output, expected2)
	}
}

func TestStreamingParser_Reset(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	// Process a JSON object
	json1 := `{"name":"John"}`

	for _, char := range json1 {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Reset the parser
	sp.Reset()

	// Process another JSON object
	json2 := `{"age":30}`

	for _, char := range json2 {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Check the result after reset
	expected := map[string]any{
		"age": int64(30),
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Unexpected result after reset. Got %v, expected %v", output, expected)
	}
}

func TestStreamingParser_Append(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	// Test the Append method with chunks
	chunks := []string{
		`{"name":`,
		`"John",`,
		`"age":`,
		`30}`,
	}

	for _, chunk := range chunks {
		err := sp.ProcessString(chunk)
		if err != nil {
			t.Fatalf("Error appending chunk '%s': %v", chunk, err)
		}
	}

	// Check the result
	expected := map[string]any{
		"name": "John",
		"age":  int64(30),
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Unexpected result. Got %v, expected %v", output, expected)
	}
}

func TestStreamingParser_StringEscapes(t *testing.T) {
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	// Process JSON with escaped characters
	json := `{"escaped":"hello\\world\\\"with quotes\\\""}`

	for _, char := range json {
		err := sp.ProcessChar(string(char))
		if err != nil {
			t.Fatalf("Error processing character '%c': %v", char, err)
		}
	}

	// Check the result
	expected := map[string]any{
		"escaped": `hello\world\"with quotes\"`,
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Unexpected result. Got %v, expected %v", output, expected)
	}
}

func TestStreamingParser_RequirementExample(t *testing.T) {
	// Test the exact example from the requirements
	output := make(map[string]any)
	sp := NewStreamingParser(&output)

	jsonStr := `{"name":"John Doe","age":30,"email":"johndoe@example.com"}`

	// Process character by character with the Append method
	for _, char := range jsonStr {
		err := sp.ProcessString(string(char))
		if err != nil {
			t.Fatalf("Error in example: %v", err)
		}
	}

	// Check the result
	expected := map[string]any{
		"name":  "John Doe",
		"age":   int64(30),
		"email": "johndoe@example.com",
	}

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("Example failed. Got %v, expected %v", output, expected)
	}
}
