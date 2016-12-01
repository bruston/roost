package lexer

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
)

type TokenType int

const eof rune = -1

const (
	EOF TokenType = iota
	Word
	Number
	String
	Colon
	Semicolon
	BracketOpen
	BracketClose
	BraceOpen
	BraceClose
	// Keywords
	Var
	If
	Else
	Then
	For
	End
)

type Token struct {
	Type     TokenType
	Value    string
	Position int
}

type Scanner struct {
	src          *bufio.Reader
	current      Token
	buf          *bytes.Buffer
	position     int
	lastRuneSize int // there is only ever need to read then unread one rune
	err          error
}

func NewScanner(r io.Reader) *Scanner {
	s := &Scanner{
		src: bufio.NewReader(r),
		buf: bytes.NewBuffer(make([]byte, 0, 1024)),
	}
	s.current = s.next()
	return s
}

func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

func (s *Scanner) Scan() bool {
	return s.current.Type != EOF
}

func (s *Scanner) Token() Token {
	t := s.current
	s.current = s.next()
	return t
}

func (s *Scanner) read() rune {
	ch, n, err := s.src.ReadRune()
	if err != nil {
		s.err = err
		return eof
	}
	s.position += n
	s.lastRuneSize = n
	return ch
}

func (s *Scanner) unread() {
	s.src.UnreadRune()
	s.position -= s.lastRuneSize
}

func (s *Scanner) peek() rune {
	ch, _, err := s.src.ReadRune()
	if err != nil {
		return eof
	}
	s.src.UnreadRune()
	return ch
}

func (s *Scanner) skipSpace() {
	for {
		ch := s.read()
		if !isWhitespace(ch) {
			s.unread()
			break
		}
	}
}

func isWhitespace(ch rune) bool { return ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' }

func (s *Scanner) scanWord() Token {
	start := s.position
	for {
		ch := s.read()
		if isWhitespace(ch) || ch == eof {
			s.unread()
			break
		}
		s.buf.WriteRune(ch)
	}
	ident := s.buf.String()
	s.buf.Reset()
	switch ident {
	case "var":
		return Token{Var, ident, start}
	case "if":
		return Token{If, ident, start}
	case "then":
		return Token{Then, ident, start}
	case "else":
		return Token{Else, ident, start}
	case "for":
		return Token{For, ident, start}
	case "end":
		return Token{End, ident, start}
	}
	return Token{Word, ident, s.position - len(ident)}
}

func (s *Scanner) scanNumber() string {
	for {
		ch := s.read()
		if !unicode.IsDigit(ch) && ch != '.' && ch != '-' {
			s.unread()
			break
		}
		s.buf.WriteRune(ch)
	}
	ident := s.buf.String()
	s.buf.Reset()

	return ident
}

func (s *Scanner) scanString() string {
	for {
		ch := s.read()
		if ch == '"' || ch == eof {
			break
		}
		s.buf.WriteRune(ch)
	}
	str := s.buf.String()
	s.buf.Reset()
	return str
}

func (s *Scanner) next() Token {
	s.skipSpace()
	peek := s.peek()
	if peek == eof {
		return Token{Type: EOF}
	}
	start := s.position
	switch peek {
	case ':':
		s.read()
		return Token{Colon, ":", start}
	case ';':
		s.read()
		return Token{Semicolon, ";", start}
	case '"':
		s.read()
		return Token{String, s.scanString(), start}
	case '[':
		s.read()
		return Token{BracketOpen, "[", start}
	case ']':
		s.read()
		return Token{BracketClose, "]", start}
	case '{':
		s.read()
		return Token{BraceOpen, "{", start}
	case '}':
		s.read()
		return Token{BraceClose, "}", start}
	}
	if unicode.IsDigit(peek) || peek == '-' {
		tok := Token{Number, s.scanNumber(), start}
		if tok.Value == "-" {
			tok.Type = Word
		}
		return tok
	}
	return s.scanWord()
}
