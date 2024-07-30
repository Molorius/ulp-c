package asm

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Molorius/ulp-c/pkg/asm/token"
)

type FileRef struct {
	Filename string
	Line     int
	Index    int // the position within the line
}

func (f FileRef) String() string {
	return fmt.Sprintf("%s:%d:%d", f.Filename, f.Line, f.Index)
}

type Token struct {
	TokenType token.Type // the type that this token represents
	Lexeme    string     // the original string
	Ref       FileRef    // the reference
	Number    int        // the number, if applicable
}

func (t Token) String() string {
	var start string
	switch t.TokenType {
	case token.Identifier:
		start = t.Lexeme
	case token.Number:
		start = fmt.Sprintf("%d", t.Number)
	case token.Unknown:
		start = fmt.Sprintf("UNKNOWN(%s)", t.Lexeme)
	default:
		start = t.TokenType.String()
	}
	return start
}

func (t *Token) Equal(other *Token) bool {
	if t.TokenType != other.TokenType {
		return false
	}
	switch t.TokenType {
	case token.Identifier:
		return t.Lexeme == other.Lexeme
	case token.Number:
		return t.Number == other.Number
	case token.Unknown:
		return t.Lexeme == other.Lexeme
	default:
		return true
	}
}

type scanner struct {
	filename     string
	line         int
	linePosition int
	position     int // position of pointer within file
	content      string
}

func (s *scanner) scanFile(content string, name string) ([]Token, error) {
	s.filename = name
	s.line = 1
	s.linePosition = 1
	s.position = 0
	s.content = content
	errs := error(nil)

	tokens := make([]Token, 0)
	for {
		t, err := s.nextToken()
		if err != nil {
			errs = errors.Join(errs, err)
		}
		tokens = append(tokens, t)
		if t.TokenType == token.EndOfFile {
			break
		}
	}
	if errs != nil {
		errs = errors.Join(fmt.Errorf("error while scanning assembly"), errs)
	}
	return tokens, errs
}

func (s *scanner) nextToken() (Token, error) {
	s.trimWhitespace()
	lexeme, ref := s.nextLexeme()
	tok, err := s.buildToken(lexeme, ref)
	return tok, err
}

func (s *scanner) buildFileRef() FileRef {
	return FileRef{Filename: s.filename, Line: s.line, Index: s.linePosition}
}

func (s *scanner) buildToken(lexeme string, ref FileRef) (Token, error) {
	tok := Token{
		TokenType: token.Unknown,
		Lexeme:    lexeme,
		Ref:       ref,
	}
	if lexeme == "" {
		tok.TokenType = token.EndOfFile
		return tok, nil
	}
	t := token.ToType(lexeme)
	if t != token.Unknown {
		tok.TokenType = t
		return tok, nil
	}

	n, err := strconv.ParseInt(lexeme, 0, 64)
	if err == nil {
		tok.TokenType = token.Number
		tok.Number = int(n)
		return tok, nil
	}

	c := lexeme[0]
	if s.isIdentifierByte(c) && !s.isNumberByte(c) && c != '.' {
		tok.TokenType = token.Identifier
		return tok, nil
	}
	return tok, UnknownTokenError{tok}
}

func (s *scanner) nextLexeme() (string, FileRef) {
	lexeme := ""
	s.trimWhitespace()
	f := s.buildFileRef()
	c, eof := s.peak()
	if eof {
		return "", f
	}
	if !s.isIdentifierByte(c) {
		s.advancePointer()
		// check if we have a "//" or "/*" comment
		if c == '/' {
			c, _ := s.peak()
			if c == '/' {
				s.skipLine()
				return s.nextLexeme()
			}
			if c == '*' {
				s.skipComment()
				return s.nextLexeme()
			}
		}
		// check if we have a "#" comment
		if c == '#' {
			s.skipLine()
			return s.nextLexeme()
		}
		return string(c), f
	}
	for {
		c, eof := s.peak()
		if eof || !s.isIdentifierByte(c) {
			return lexeme, f
		}
		lexeme += string(c)
		s.advancePointer()
	}
}

func (s *scanner) skipLine() {
	for {
		c, eof := s.peak()
		if eof {
			return
		}
		s.advancePointer()
		if c == '\n' {
			return
		}
	}
}

func (s *scanner) skipComment() {
	for {
		c, eof := s.peak()
		if eof {
			return
		}
		s.advancePointer()
		if c == '*' {
			c, eof := s.peak()
			if eof {
				return
			}
			s.advancePointer()
			if c == '/' {
				return
			}
		}
	}
}

func (s *scanner) isIdentifierByte(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		s.isNumberByte(b) ||
		b == '_' || b == '.'
}

func (s *scanner) isNumberByte(b byte) bool {
	return (b >= '0' && b <= '9')
}

// Gets the next character.
// Returns the character and true if end of file.
func (s *scanner) peak() (byte, bool) {
	if s.position >= len(s.content) {
		return 0, true
	}
	return s.content[s.position], false
}

func (s *scanner) trimWhitespace() {
	for {
		c, eof := s.peak()
		if eof || !s.isWhitespace(c) {
			break
		}
		s.advancePointer()
	}
}

func (s *scanner) isWhitespace(b byte) bool {
	return b == ' ' || b == '\n' || b == '\r' || b == '\t'
}

func (s *scanner) advancePointer() {
	c, eof := s.peak()
	if eof {
		return
	}
	if c == '\n' {
		s.line += 1
		s.linePosition = 0
	}
	s.linePosition += 1
	s.position += 1
}
