# FlexJSON: A Robust Partial and Streaming JSON Parser for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/jpoz/flexjson.svg)](https://pkg.go.dev/github.com/jpoz/flexjson)
[![Go Report Card](https://goreportcard.com/badge/github.com/jpoz/flexjson)](https://goreportcard.com/report/github.com/jpoz/flexjson)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

FlexJSON parses incomplete or streaming JSON data. Unlike standard JSON parsers that require valid, complete JSON input, FlexJSON gracefully handles partial JSON fragments and streams of characters, extracting as much structured data as possible.

## 🌟 Features

- **Partial JSON Parsing**: Extract data from incomplete JSON fragments
  - `{"key": 123` → `map[string]any{"key": 123}`
  - `{"key": 1234, "key2":` → `map[string]any{"key": 1234, "key2": nil}`

- **Character-by-Character Streaming**: Process JSON one character at a time
  - Ideal for network streams, telemetry data, or large files
  - Updates an output map in real-time as data arrives

- **Nested Structure Support**: Handles complex nested objects and arrays
  - Properly tracks hierarchy in deeply nested structures
  - Maintains context across partial fragments

- **Resilient Parsing**: Recovers gracefully from unexpected input
  - No panic on malformed input
  - Extracts maximum valid data even from corrupted JSON

- **Zero Dependencies**: Pure Go implementation with no external dependencies

## 📦 Installation

```bash
go get github.com/jpoz/flexjson
```

## 🚀 Quick Start

### Parsing Partial JSON

```go
package main

import (
    "fmt"
    "github.com/jpoz/flexjson"
)

func main() {
    // Parse incomplete JSON
    partialJSON := `{"name": "John", "age": 30, "city":`
    
    result, err := flexjson.ParsePartialJSONObject(partialJSON)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("Parsed result: %v\n", result)
    // Output: Parsed result: map[name:John age:30 city:<nil>]
}
```

### Streaming JSON Character-by-Character

```go
package main

import (
    "fmt"
    "github.com/jpoz/flexjson"
)

func main() {
    // Example JSON string
    jsonStrs := []string{`{"name":"John Doe"`, `,"age":30`, `,"email":"johndoe@example.com"}`}
    
    // Create output map
    output := map[string]any{}
    
    // Create streaming parser
    sp := flexjson.NewStreamingParser(&output)
    
    // Process each character
    for _, str := range jsonStrs {
        err := sp.Append(str)
        if err != nil {
            fmt.Printf("Error: %v\n", err)
            return
        }
        
        // The output map is updated after each character
        fmt.Printf("Current state: %v\n", output)
    }
    
    fmt.Printf("Final result: %v\n", output)
}
```

## ⚙️ How It Works

FlexJSON uses a custom lexer and parser system for the partial JSON parsing, and a state machine approach for streaming parsing:

1. **Lexer**: Tokenizes the input string into JSON tokens (strings, numbers, booleans, etc.)
2. **Parser**: Converts tokens into a structured map representation
3. **StreamingParser**: Maintains stacks of containers and keys to track position in the JSON hierarchy

The library intelligently handles incomplete input by:
- Treating unexpected EOF as valid termination
- Providing default values (nil) for incomplete key-value pairs
- Maintaining context across nested structures

## 🧪 Testing

The library includes comprehensive test coverage for both partial and streaming parsing:

```bash
go test -v github.com/jpoz/flexjson
```
