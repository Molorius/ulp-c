package hlp

import "fmt"

type UnknownTokenError struct {
	token Token
}

func (e UnknownTokenError) Error() string {
	return fmt.Sprintf("%s: unknown token \"%s\"", e.token.Ref, e.token.Lexeme)
}
