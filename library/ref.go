package library

import (
	"strconv"
	"strings"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

const (
	codeBadReference    = "bad-reference"
	codeUnknownFunction = "unknown-function"
	codeVersionMissing  = "version-missing"
	codeVersionNotFound = "version-not-found"
	codeVersionStopped  = "version-stopped"
	maxNameLength       = 40
	maxVersionDigits    = 9
)

type Ref struct {
	Library string
	Name    string
	Version int
}

func (r Ref) String() string {
	text := r.Library + "." + r.Name
	if r.Version == 0 {
		return text
	}
	return text + "@" + strconv.Itoa(r.Version)
}

func (r Ref) function() string { return r.Library + "." + r.Name }

func ParseRef(text string) (Ref, *formula.Problem) {
	reject := func() (Ref, *formula.Problem) {
		return Ref{}, &formula.Problem{
			Code:     codeBadReference,
			Sentence: "`" + text + "` is not written as library.function@N, with N a whole number from 1.",
		}
	}
	parts := strings.Split(text, "@")
	if len(parts) > 2 {
		return reject()
	}
	names := strings.Split(parts[0], ".")
	if len(names) != 2 || !validName(names[0]) || !validName(names[1]) {
		return reject()
	}
	ref := Ref{Library: names[0], Name: names[1]}
	if len(parts) == 2 {
		version, ok := parseVersion(parts[1])
		if !ok {
			return reject()
		}
		ref.Version = version
	}
	return ref, nil
}

func parseVersion(text string) (int, bool) {
	if text == "" || len(text) > maxVersionDigits || text[0] == '0' {
		return 0, false
	}
	for _, r := range text {
		if r < '0' || r > '9' {
			return 0, false
		}
	}
	version, err := strconv.Atoi(text)
	return version, err == nil
}

func validName(name string) bool {
	if len(name) == 0 || len(name) > maxNameLength || name[0] < 'a' || name[0] > 'z' {
		return false
	}
	for _, r := range name {
		letter := r >= 'a' && r <= 'z'
		digit := r >= '0' && r <= '9'
		if !letter && !digit && r != '_' {
			return false
		}
	}
	return true
}
