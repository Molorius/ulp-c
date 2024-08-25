/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package hlp

import "fmt"

type UnknownTokenError struct {
	token Token
}

func (e UnknownTokenError) Error() string {
	return fmt.Sprintf("%s: unknown token \"%s\"", e.token.Ref, e.token.Lexeme)
}