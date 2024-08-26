/*
Copyright 2024 Blake Felt blake.w.felt@gmail.com

This Source Code Form is subject to the terms of the Mozilla Public
License, v. 2.0. If a copy of the MPL was not distributed with this
file, You can obtain one at https://mozilla.org/MPL/2.0/.
*/
package hlp

import (
	"errors"
	"fmt"
)

type Hlp struct {
}

type HlpFile struct {
	Name     string
	Contents string
}

func (h *Hlp) Build(files []HlpFile) error {
	// scan files to create token stream
	t := make([]Token, 0)
	s := Scanner{}
	errs := error(nil)
	for _, f := range files {
		stream, err := s.ScanFile(f.Contents, f.Name)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		t = append(t, stream...)
	}
	if errs != nil {
		return errs
	}
	fmt.Println(t)

	// parse it
	p := parser{}
	stmnts, err := p.parseTokens(t)
	if err != nil {
		return errs
	}

	fmt.Println(stmnts)
	return nil
}
