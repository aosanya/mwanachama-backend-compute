package library

import (
	"errors"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

var (
	ErrBadRelease       = errors.New("the release is not complete")
	ErrOutOfOrder       = errors.New("a version starts at 1 and goes up by 1")
	ErrAlreadyPublished = errors.New("a published version is never edited")
	ErrNoteIncomplete   = errors.New("every version from 2 carries a note with three parts")
	ErrBadMark          = errors.New("the known-wrong mark is refused")
)

type Example struct {
	Input string
	Old   string
	New   string
}

type Note struct {
	Changed string
	Matters string
	Example Example
}

type Implementation func(args []formula.Value) (formula.Value, *formula.Problem)

type Release struct {
	Library   string
	Name      string
	Version   int
	Signature formula.Signature
	Apply     Implementation
	Note      Note
}

func (r Release) ref() Ref {
	return Ref{Library: r.Library, Name: r.Name, Version: r.Version}
}

func (n Note) complete() bool {
	for _, part := range []string{n.Changed, n.Matters, n.Example.Input, n.Example.Old, n.Example.New} {
		if strings.TrimSpace(part) == "" {
			return false
		}
	}
	return true
}

func copySignature(s formula.Signature) formula.Signature {
	s.Params = append([]formula.Param{}, s.Params...)
	return s
}

func validRelease(r Release) bool {
	return validName(r.Library) && validName(r.Name) && r.Version >= 1 && r.Apply != nil && r.Signature.Result != ""
}
