/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package hlp

type parser struct {
	tokens   []Token
	position int
}

func (p *parser) parseTokens(tokens []Token) ([]StaticStatement, error) {
	_ = tokens
	_ = p.tokens
	_ = p.position
	return nil, nil
}
