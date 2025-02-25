package flexjson

import (
	"errors"
	"strconv"
)

// Token types used by the lexer
type TokenType int

const (
	TokenError TokenType = iota
	TokenEOF
	TokenLeftBrace
	TokenRightBrace
	TokenLeftBracket
	TokenRightBracket
	TokenColon
	TokenComma
	TokenString
	TokenNumber
	TokenTrue
	TokenFalse
	TokenNull
)

// Token represents a JSON token
type Token struct {
	Type  TokenType
	Value string
}

// Lexer tokenizes JSON input
type Lexer struct {
	input  string
	pos    int
	start  int
	tokens []Token
}

// NewLexer creates a new JSON lexer
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		start:  0,
		tokens: []Token{},
	}
}

// Tokenize converts the input string into tokens
func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		l.start = l.pos
		l.scanToken()
	}

	// Add EOF token
	l.tokens = append(l.tokens, Token{Type: TokenEOF})
	return l.tokens
}

// scanToken scans the next token
func (l *Lexer) scanToken() {
	// Check if we're at the end of input
	if l.pos >= len(l.input) {
		return
	}

	c := l.input[l.pos]

	switch c {
	case '{':
		l.addToken(TokenLeftBrace)
	case '}':
		l.addToken(TokenRightBrace)
	case '[':
		l.addToken(TokenLeftBracket)
	case ']':
		l.addToken(TokenRightBracket)
	case ':':
		l.addToken(TokenColon)
	case ',':
		l.addToken(TokenComma)
	case '"':
		l.scanString()
	case ' ', '\t', '\r', '\n':
		// Skip whitespace
		l.pos++
	default:
		if isDigit(c) || c == '-' {
			l.scanNumber()
		} else if isAlpha(c) {
			l.scanIdentifier()
		} else {
			// Skip unknown characters
			l.pos++
		}
	}
}

// addToken adds a token to the token list
func (l *Lexer) addToken(tokenType TokenType) {
	value := string(l.input[l.pos])
	l.tokens = append(l.tokens, Token{Type: tokenType, Value: value})
	l.pos++
}

// scanString scans a string token (handling quotes and escapes)
func (l *Lexer) scanString() {
	l.pos++ // Skip opening quote

	startPos := l.pos

	// Continue until we find a closing quote or reach the end
	for l.pos < len(l.input) && l.input[l.pos] != '"' {
		if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
			l.pos++ // Skip escape character
		}
		l.pos++
	}

	value := l.input[startPos:l.pos]
	l.tokens = append(l.tokens, Token{Type: TokenString, Value: value})

	if l.pos < len(l.input) {
		l.pos++ // Skip closing quote if it exists
	}
}

// scanNumber scans a number token
func (l *Lexer) scanNumber() {
	startPos := l.pos

	// Handle minus sign
	if l.input[l.pos] == '-' {
		l.pos++
	}

	// Integer part
	for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
		l.pos++
	}

	// Fractional part
	if l.pos < len(l.input) && l.input[l.pos] == '.' {
		l.pos++
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
	}

	// Exponent part
	if l.pos < len(l.input) && (l.input[l.pos] == 'e' || l.input[l.pos] == 'E') {
		l.pos++
		if l.pos < len(l.input) && (l.input[l.pos] == '+' || l.input[l.pos] == '-') {
			l.pos++
		}
		for l.pos < len(l.input) && isDigit(l.input[l.pos]) {
			l.pos++
		}
	}

	value := l.input[startPos:l.pos]
	l.tokens = append(l.tokens, Token{Type: TokenNumber, Value: value})
}

// scanIdentifier scans identifiers like true, false, null
func (l *Lexer) scanIdentifier() {
	startPos := l.pos

	for l.pos < len(l.input) && isAlphaNumeric(l.input[l.pos]) {
		l.pos++
	}

	value := l.input[startPos:l.pos]

	// Check which identifier it is
	switch value {
	case "true":
		l.tokens = append(l.tokens, Token{Type: TokenTrue, Value: value})
	case "false":
		l.tokens = append(l.tokens, Token{Type: TokenFalse, Value: value})
	case "null":
		l.tokens = append(l.tokens, Token{Type: TokenNull, Value: value})
	default:
		// Skip unknown identifiers
	}
}

// Helper functions
func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isAlpha(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isAlphaNumeric(c byte) bool {
	return isAlpha(c) || isDigit(c)
}

// Parser parses tokens into a JSON value
type Parser struct {
	tokens  []Token
	current int
}

// NewParser creates a new JSON parser
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens:  tokens,
		current: 0,
	}
}

// Parse parses tokens into a JSON value
func (p *Parser) Parse() (interface{}, error) {
	if len(p.tokens) == 0 {
		return nil, errors.New("no tokens to parse")
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}
	return value, nil
}

// parseValue parses any JSON value
func (p *Parser) parseValue() (interface{}, error) {
	if p.isAtEnd() {
		return nil, errors.New("unexpected end of JSON")
	}

	token := p.peek()

	switch token.Type {
	case TokenLeftBrace:
		return p.parseObject()
	case TokenLeftBracket:
		return p.parseArray()
	case TokenString:
		p.advance()
		return token.Value, nil
	case TokenNumber:
		p.advance()
		// Try parsing as int first
		if i, err := strconv.ParseInt(token.Value, 10, 64); err == nil {
			return i, nil
		}
		// Try parsing as float
		if f, err := strconv.ParseFloat(token.Value, 64); err == nil {
			return f, nil
		}
		return nil, errors.New("invalid number: " + token.Value)
	case TokenTrue:
		p.advance()
		return true, nil
	case TokenFalse:
		p.advance()
		return false, nil
	case TokenNull:
		p.advance()
		return nil, nil
	case TokenEOF:
		return nil, errors.New("unexpected end of JSON")
	default:
		p.advance()
		return nil, errors.New("unexpected token: " + token.Value)
	}
}

// parseObject parses a JSON object, handling incomplete objects
func (p *Parser) parseObject() (map[string]interface{}, error) {
	obj := make(map[string]interface{})

	// Consume the left brace
	p.advance()

	// Handle empty object
	if p.check(TokenRightBrace) {
		p.advance()
		return obj, nil
	}

	for {
		// End of input - return partial object
		if p.isAtEnd() {
			return obj, nil
		}

		// We need a string key
		if !p.check(TokenString) {
			// If we don't have a string key but we have EOF, return what we have
			if p.check(TokenEOF) {
				return obj, nil
			}
			return nil, errors.New("expected string key in object")
		}

		// Get the key
		key := p.peek().Value
		p.advance()

		// We need a colon
		if !p.check(TokenColon) {
			// If we don't have a colon but we have EOF, set value to nil and return
			if p.check(TokenEOF) {
				obj[key] = nil
				return obj, nil
			}
			return nil, errors.New("expected ':' after key in object")
		}

		// Consume the colon
		p.advance()

		// Handle EOF after colon
		if p.check(TokenEOF) {
			obj[key] = nil
			return obj, nil
		}

		// Parse the value
		value, err := p.parseValue()
		if err != nil {
			// If we have an error and we're at EOF, just set to nil and return
			if p.check(TokenEOF) {
				obj[key] = nil
				return obj, nil
			}
			return nil, err
		}

		// Add the key-value pair
		obj[key] = value

		// Check for comma or right brace
		if !p.check(TokenComma) && !p.check(TokenRightBrace) {
			// If we don't have a comma or right brace but have EOF, return what we have
			if p.check(TokenEOF) {
				return obj, nil
			}
			return nil, errors.New("expected ',' or '}' after object value")
		}

		// If we're at the end of the object, we're done
		if p.check(TokenRightBrace) {
			p.advance()
			return obj, nil
		}

		// Consume the comma
		p.advance()

		// Handle trailing comma at EOF
		if p.check(TokenEOF) {
			return obj, nil
		}
	}
}

// parseArray parses a JSON array, handling incomplete arrays
func (p *Parser) parseArray() ([]interface{}, error) {
	arr := make([]interface{}, 0)

	// Consume the left bracket
	p.advance()

	// Handle empty array
	if p.check(TokenRightBracket) {
		p.advance()
		return arr, nil
	}

	for {
		// End of input - return partial array
		if p.isAtEnd() {
			return arr, nil
		}

		// Parse the value
		value, err := p.parseValue()
		if err != nil {
			// If we have an error but we're at EOF, return what we have
			if p.check(TokenEOF) {
				return arr, nil
			}
			return nil, err
		}

		// Add the value
		arr = append(arr, value)

		// Check for comma or right bracket
		if !p.check(TokenComma) && !p.check(TokenRightBracket) {
			// If we don't have a comma or right bracket but have EOF, return what we have
			if p.check(TokenEOF) {
				return arr, nil
			}
			return nil, errors.New("expected ',' or ']' after array value")
		}

		// If we're at the end of the array, we're done
		if p.check(TokenRightBracket) {
			p.advance()
			return arr, nil
		}

		// Consume the comma
		p.advance()

		// Handle trailing comma at EOF
		if p.check(TokenEOF) {
			return arr, nil
		}
	}
}

// Helper methods for parser
func (p *Parser) advance() {
	if !p.isAtEnd() {
		p.current++
	}
}

func (p *Parser) peek() Token {
	return p.tokens[p.current]
}

func (p *Parser) check(tokenType TokenType) bool {
	if p.isAtEnd() {
		return tokenType == TokenEOF
	}
	return p.peek().Type == tokenType
}

func (p *Parser) isAtEnd() bool {
	return p.current >= len(p.tokens) || p.tokens[p.current].Type == TokenEOF
}

// ParsePartialJSON parses a partial JSON string into a Go value
func ParsePartialJSON(input string) (interface{}, error) {
	lexer := NewLexer(input)
	tokens := lexer.Tokenize()

	parser := NewParser(tokens)
	return parser.Parse()
}

// ParsePartialJSONObject parses a partial JSON string into a map[string]any
// This is the main function that should be used by clients
func ParsePartialJSONObject(input string) (map[string]any, error) {
	result, err := ParsePartialJSON(input)
	if err != nil {
		return nil, err
	}

	// If result is already a map, return it
	if obj, ok := result.(map[string]interface{}); ok {
		// In Go 1.18+, map[string]any is the same as map[string]interface{}
		return obj, nil
	}

	// If result is something else, return an error
	return nil, errors.New("input is not a JSON object")
}
