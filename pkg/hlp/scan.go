package hlp

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/Molorius/ulp-c/pkg/hlp/token"
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
	TokenType token.Type
	Lexeme    string
	Ref       FileRef
	Number    int // the number, if applicable
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

type Scanner struct {
	filename     string
	line         int
	linePosition int
	position     int // position of pointer within file
	content      string
}

func (s *Scanner) ScanFile(content string, name string) ([]Token, error) {
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
		errs = errors.Join(fmt.Errorf("error while scanning hlp"), errs)
	}
	return tokens, errs
}

func (s *Scanner) nextToken() (Token, error) {
	s.trimWhitespace()
	lexeme, ref := s.nextLexeme()
	tok, err := s.buildToken(lexeme, ref)
	return tok, err
}

func (s *Scanner) buildToken(lexeme string, ref FileRef) (Token, error) {
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

func (s *Scanner) buildFileRef() FileRef {
	return FileRef{Filename: s.filename, Line: s.line, Index: s.linePosition}
}

// this is a map for fast access, we only use the key
var multiByte = map[string]int{
	"<<": 0,
	"<=": 0,
	">=": 0,
	">>": 0,
	"!=": 0,
	"==": 0,
}

func (s *Scanner) nextLexeme() (string, FileRef) {
	lexeme := ""
	s.trimWhitespace()
	f := s.buildFileRef()
	c, eof := s.peak()
	if eof {
		return "", f
	}
	if !s.isIdentifierByte(c) {
		s.advancePointer()
		next, _ := s.peak()
		// check if we have a "//"
		if c == '/' {
			next, _ := s.peak()
			if next == '/' {
				s.skipLine()
				return s.nextLexeme()
			}
		}
		// check if it's a multibyte token
		two := string(c) + string(next)
		_, ok := multiByte[two]
		if ok {
			s.advancePointer()
			return two, f
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

func (s *Scanner) skipLine() {
	for {
		c, eof := s.peak()
		if eof || c == '\n' {
			return
		}
		s.advancePointer()
	}
}

func (s *Scanner) isIdentifierByte(b byte) bool {
	return (b >= 'a' && b <= 'z') ||
		(b >= 'A' && b <= 'Z') ||
		s.isNumberByte(b) ||
		b == '_' || b == '.'
}

func (s *Scanner) isNumberByte(b byte) bool {
	return (b >= '0' && b <= '9')
}

func (s *Scanner) trimWhitespace() {
	for {
		c, eof := s.peak()
		if eof || !s.isWhitespace(c) {
			break
		}
		s.advancePointer()
	}
}

func (s *Scanner) isWhitespace(b byte) bool {
	return b == ' ' || b == '\r' || b == '\t'
}

func (s *Scanner) peak() (byte, bool) {
	if s.position >= len(s.content) {
		return 0, true
	}
	return s.content[s.position], false
}

func (s *Scanner) advancePointer() {
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
