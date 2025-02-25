package flexjson

import (
	"errors"
	"fmt"
	"strconv"
)

// StreamingParser is a simplified JSON parser that processes JSON character by character
// and updates an output map as it goes along.
type StreamingParser struct {
	output       *map[string]any // Pointer to the output map
	stack        []interface{}   // Stack of containers (maps/slices)
	keys         []string        // Stack of keys
	paths        []string        // Current path in the JSON
	buffer       string          // Buffer for the current token
	isEscaping   bool            // Whether we're currently escaping a character
	inString     bool            // Whether we're currently inside a string
	expectingKey bool            // Whether we're expecting a key
	expectColon  bool            // Whether we're expecting a colon
	lastChar     string          // Last processed character
	debug        bool            // Whether to print debug messages
}

// NewStreamingParser creates a new StreamingParser that will update the provided map
func NewStreamingParser(output *map[string]any) *StreamingParser {
	if output == nil {
		m := make(map[string]any)
		output = &m
	}

	// Clear the output map to start fresh
	for k := range *output {
		delete(*output, k)
	}

	return &StreamingParser{
		output:       output,
		stack:        []interface{}{output},
		keys:         []string{},
		paths:        []string{},
		buffer:       "",
		isEscaping:   false,
		inString:     false,
		expectingKey: true,
		expectColon:  false,
		lastChar:     "",
	}
}

// ProcessString processes a chunk of JSON data character by character
func (sp *StreamingParser) ProcessString(chunk string) error {
	for _, c := range chunk {
		err := sp.ProcessChar(string(c))
		if err != nil {
			return err
		}
	}
	return nil
}

// ProcessChar processes a single character in the JSON stream
func (sp *StreamingParser) ProcessChar(c string) error {
	sp.log("- %s\texpecting key: %v, expecting colon: %v, isEscaping: %v, inString: %v, buffer: %s\n", c,
		sp.expectingKey, sp.expectColon, sp.isEscaping, sp.inString, sp.buffer)

	if (c == "," || c == "}" || c == "]") && sp.buffer != "" {
		// Try to parse as a number
		if value, err := sp.parseNumber(); err == nil {
			sp.log("\tAdding number value: %v\n", value)
			sp.addValue(value)
			sp.buffer = ""
		}
	}

	// Handle string state (special handling for escaping)
	if sp.inString {
		if sp.isEscaping {
			// We're currently escaping
			sp.log("\tEscaping character\n")
			sp.buffer += c
			sp.isEscaping = false
			sp.lastChar = c
			return nil
		}

		if c == "\\" {
			sp.log("\tStart of escaping character\n")
			sp.isEscaping = true
			sp.lastChar = c
			return nil
		}

		if c == "\"" {
			sp.log("End of string\n")
			// End of string
			sp.inString = false

			// Handle differently based on context
			if sp.expectingKey {
				sp.log("\tStoring as key\n")
				// We just parsed a key
				sp.keys = append(sp.keys, sp.buffer)
				sp.expectingKey = false
				sp.expectColon = true
			} else {
				sp.log("\tAdding as value\n")
				// We just parsed a string value
				sp.addValue(sp.buffer)
			}

			sp.buffer = ""
			sp.lastChar = c
			return nil
		}

		// Regular character in string
		sp.buffer += c
		sp.lastChar = c
		return nil
	}

	// Handle other states
	switch c {
	case " ", "\t", "\r", "\n":
		sp.log("Whitespace\n")
		// Skip whitespace
		sp.lastChar = c
		return nil

	case "{":
		sp.log("Start of object\n")
		// Start of an object
		if len(sp.stack) == 1 && len(sp.keys) == 0 {
			// Root object - already setup in our output
			sp.log("\tRoot object\n")
			sp.expectingKey = true
			sp.lastChar = c
			return nil
		}

		sp.log("\tCreating new object\n")
		// Create new object
		newObj := make(map[string]any)

		// Add it to its parent
		sp.addValue(newObj)

		// Push it onto the stack
		sp.stack = append(sp.stack, newObj)
		sp.expectingKey = true
		sp.lastChar = c
		return nil

	case "}":
		sp.log("End of object\n")
		// End of an object
		if len(sp.stack) > 1 {
			sp.stack = sp.stack[:len(sp.stack)-1] // Pop from stack

			// If we have keys, also pop the last key
			if len(sp.keys) > 0 {
				sp.keys = sp.keys[:len(sp.keys)-1]
			}
		}
		sp.expectingKey = false
		sp.expectColon = false
		sp.lastChar = c
		return nil

	case "[":
		sp.log("Start of array\n")
		// Start of an array
		newArray := make([]interface{}, 0)

		// Add it to its parent
		sp.addValue(&newArray)

		// Push it onto the stack
		sp.stack = append(sp.stack, &newArray)
		sp.expectingKey = false
		sp.lastChar = c
		return nil

	case "]":
		sp.log("End of array")
		// End of an array
		if len(sp.stack) > 1 {
			sp.stack = sp.stack[:len(sp.stack)-1] // Pop from stack

			// If we have keys, also pop the last key
			if len(sp.keys) > 0 {
				sp.keys = sp.keys[:len(sp.keys)-1]
			}
		}
		sp.expectingKey = false
		sp.expectColon = false
		sp.lastChar = c
		return nil

	case "\"":
		sp.log("Start of string\n")
		// Start of a string
		sp.inString = true
		sp.buffer = ""
		sp.lastChar = c
		return nil

	case ":":
		sp.log("Colon. Expecting: %#v\n", sp.expectColon)
		// Colon after key
		if !sp.expectColon {
			return errors.New(
				fmt.Sprintf("unexpected ':' - state: %#v", sp),
			)
		}
		sp.expectColon = false
		sp.lastChar = c
		return nil

	case ",":
		sp.log("Comma\n")
		// Comma between values or key-value pairs
		// After a comma, if the parent is an object, we expect a key
		if parent, ok := sp.getCurrentContainer(); ok {
			switch parent.(type) {
			case *map[string]any, map[string]any:
				sp.log("\tParent is an object. Expecting key\n")
				sp.expectingKey = true
			case *[]interface{}:
				sp.log("\tParent is an array. Not expecting key\n")
				sp.expectingKey = false
			default:
				sp.log("\tWarning: Parent is not an object or array. Not expecting key. Parent: %#v\n", parent)
			}
		}
		sp.lastChar = c
		return nil

	case "t":
		// Start of 'true'
		if sp.buffer != "" {
			return errors.New("unexpected 't'")
		}
		sp.buffer = "t"
		sp.lastChar = c
		return nil

	case "r":
		// Part of 'true'
		if sp.buffer == "t" {
			sp.buffer = "tr"
			sp.lastChar = c
			return nil
		}
		return errors.New("unexpected 'r'")

	case "u":
		// Part of 'true'
		if sp.buffer == "tr" {
			sp.buffer = "tru"
			sp.lastChar = c
			return nil
		}

		// Part of 'null'
		if sp.buffer == "n" {
			sp.buffer = "nu"
			sp.lastChar = c
			return nil
		}

		return errors.New("unexpected 'u'")

	case "e":
		// End of 'true' or part of 'false'
		if sp.buffer == "tru" {
			// Complete 'true'
			sp.addValue(true)
			sp.buffer = ""
			sp.lastChar = c
			return nil
		}
		if sp.buffer == "fals" {
			// Complete 'false'
			sp.addValue(false)
			sp.buffer = ""
			sp.lastChar = c
			return nil
		}
		return errors.New("unexpected 'e'")

	case "f":
		// Start of 'false'
		if sp.buffer != "" {
			return errors.New("unexpected 'f'")
		}
		sp.buffer = "f"
		sp.lastChar = c
		return nil

	case "a":
		// Part of 'false'
		if sp.buffer == "f" {
			sp.buffer = "fa"
			sp.lastChar = c
			return nil
		}
		return errors.New("unexpected 'a'")

	case "l":
		// Part of 'false'
		if sp.buffer == "fa" {
			sp.buffer = "fal"
			sp.lastChar = c
			return nil
		}

		// Part of 'null'
		if sp.buffer == "nu" {
			sp.buffer = "nul"
			sp.lastChar = c
			return nil
		}
		if sp.buffer == "nul" {
			// Complete 'null'
			sp.addValue(nil)
			sp.buffer = ""
			sp.lastChar = c
			return nil
		}
		return errors.New("unexpected 'l'")
	case "s":
		// Part of 'false'
		if sp.buffer == "fal" {
			sp.buffer = "fals"
			sp.lastChar = c
			return nil
		}
		return errors.New("unexpected 's'")

	case "n":
		// Start of 'null'
		if sp.buffer != "" {
			return errors.New("unexpected 'n'")
		}
		sp.buffer = "n"
		sp.lastChar = c
		return nil
	default:
		if (c >= "0" && c <= "9") || c == "-" || c == "." || c == "+" || c == "e" || c == "E" {
			sp.buffer += c
			sp.lastChar = c
			return nil
		}

		return errors.New("unexpected character: " + c)
	}
}

// parseNumber parses the current buffer as a number
func (sp *StreamingParser) parseNumber() (interface{}, error) {
	// Try to parse as integer first
	if i, err := strconv.ParseInt(sp.buffer, 10, 64); err == nil {
		return i, nil
	}

	// Try to parse as float
	if f, err := strconv.ParseFloat(sp.buffer, 64); err == nil {
		return f, nil
	}

	return nil, errors.New("invalid number: " + sp.buffer)
}

// getCurrentContainer gets the current container (map or slice) from the stack
func (sp *StreamingParser) getCurrentContainer() (interface{}, bool) {
	if len(sp.stack) == 0 {
		return nil, false
	}
	return sp.stack[len(sp.stack)-1], true
}

// addValue adds a value to the current container
func (sp *StreamingParser) addValue(value interface{}) {
	if len(sp.stack) == 0 {
		return
	}

	current := sp.stack[len(sp.stack)-1]

	switch container := current.(type) {
	case *map[string]any:
		// Add to map with the current key
		if len(sp.keys) > 0 {
			key := sp.keys[len(sp.keys)-1]
			(*container)[key] = value

			// Don't remove the key here, it gets removed when we close the object
		}
	case map[string]any:
		// Add to map with the current key
		if len(sp.keys) > 0 {
			key := sp.keys[len(sp.keys)-1]
			container[key] = value

			// Don't remove the key here, it gets removed when we close the object
		}
	case *[]interface{}:
		// Add to slice
		*container = append(*container, value)
	case []interface{}:
		// Add to slice
		newSlice := append(container, value)

		// Update the parent with the new slice
		if len(sp.stack) >= 2 {
			parent := sp.stack[len(sp.stack)-2]

			switch p := parent.(type) {
			case *map[string]any:
				// Parent is a map
				if len(sp.keys) >= 2 {
					key := sp.keys[len(sp.keys)-2]
					(*p)[key] = newSlice
				}
			case map[string]any:
				// Parent is a map
				if len(sp.keys) >= 2 {
					key := sp.keys[len(sp.keys)-2]
					p[key] = newSlice
				}
			}
		}

		// Update the current container in the stack
		sp.stack[len(sp.stack)-1] = newSlice
	}
}

// Reset resets the parser state
func (sp *StreamingParser) Reset() {
	// Clear the output map
	for k := range *sp.output {
		delete(*sp.output, k)
	}

	// Reset parser state
	sp.stack = []interface{}{sp.output}
	sp.keys = []string{}
	sp.paths = []string{}
	sp.buffer = ""
	sp.isEscaping = false
	sp.inString = false
	sp.expectingKey = true
	sp.expectColon = false
	sp.lastChar = ""
}

func (sp *StreamingParser) SetDebug(value bool) {
	sp.debug = value
}

func (sp *StreamingParser) log(msg string, args ...interface{}) {
	if sp.debug {
		fmt.Printf(msg, args...)
	}
}

// GetCurrentOutput returns the current output map
func (sp *StreamingParser) GetCurrentOutput() map[string]any {
	return *sp.output
}
