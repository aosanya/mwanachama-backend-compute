package library

import (
	"fmt"
	"strings"
	"sync"

	"github.com/aosanya/mwanachama-backend-compute/formula"
)

type functionKey struct {
	library string
	name    string
}

type history struct {
	releases []Release
	marks    map[int]int
}

type Registry struct {
	mu        sync.RWMutex
	functions map[functionKey]*history
	stops     StopSource
}

func NewRegistry() *Registry {
	return &Registry{functions: map[functionKey]*history{}}
}

func (r *Registry) Publish(release Release) error {
	if !validRelease(release) {
		return fmt.Errorf("%w: %s needs a lower-case library and function name, a version from 1, a result type and an implementation", ErrBadRelease, release.ref())
	}
	release.Signature = copySignature(release.Signature)
	key := functionKey{release.Library, release.Name}
	r.mu.Lock()
	defer r.mu.Unlock()
	h := r.functions[key]
	published := 0
	if h != nil {
		published = len(h.releases)
	}
	if release.Version <= published {
		return fmt.Errorf("%w: %s is already published", ErrAlreadyPublished, release.ref())
	}
	if release.Version != published+1 {
		return fmt.Errorf("%w: %s comes after version %d", ErrOutOfOrder, release.ref(), published)
	}
	if release.Version >= 2 && !release.Note.complete() {
		return fmt.Errorf("%w: %s", ErrNoteIncomplete, release.ref())
	}
	if h == nil {
		h = &history{marks: map[int]int{}}
		r.functions[key] = h
	}
	h.releases = append(h.releases, release)
	return nil
}

func (r *Registry) ReadStopsFrom(source StopSource) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stops = source
}

func (r *Registry) Versions(library string, name string) []int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h := r.functions[functionKey{library, name}]
	if h == nil {
		return []int{}
	}
	versions := make([]int, len(h.releases))
	for i := range h.releases {
		versions[i] = i + 1
	}
	return versions
}

func (r *Registry) Function(library string, name string, version int) (formula.Signature, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	release, found := r.release(library, name, version)
	if !found {
		return formula.Signature{}, false
	}
	return copySignature(release.Signature), true
}

func (r *Registry) release(library string, name string, version int) (Release, bool) {
	h := r.functions[functionKey{library, name}]
	if h == nil || version < 1 || version > len(h.releases) {
		return Release{}, false
	}
	return h.releases[version-1], true
}

func (r *Registry) latest(library string, name string) int {
	if h := r.functions[functionKey{library, name}]; h != nil {
		return len(h.releases)
	}
	return 0
}

func (r *Registry) Resolve(ref Ref) (Release, *formula.Problem) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.resolveLocked(ref)
}

func (r *Registry) resolveLocked(ref Ref) (Release, *formula.Problem) {
	latest := r.latest(ref.Library, ref.Name)
	if latest == 0 {
		return Release{}, problem(codeUnknownFunction, "The function "+ref.function()+" is not known.")
	}
	if ref.Version == 0 {
		available := make([]string, latest)
		for i := range available {
			available[i] = Ref{Library: ref.Library, Name: ref.Name, Version: i + 1}.String()
		}
		return Release{}, problem(codeVersionMissing, ref.function()+" needs a version. Available: "+strings.Join(available, ", ")+".")
	}
	release, found := r.release(ref.Library, ref.Name, ref.Version)
	if !found {
		latestRef := Ref{Library: ref.Library, Name: ref.Name, Version: latest}
		return Release{}, problem(codeVersionNotFound, "`"+ref.String()+"` does not exist. Latest is `"+latestRef.String()+"`.")
	}
	release.Signature = copySignature(release.Signature)
	return release, nil
}

func (r *Registry) Apply(library string, name string, version int, args []formula.Value) (formula.Value, *formula.Problem) {
	r.mu.RLock()
	ref := Ref{Library: library, Name: name, Version: version}
	release, resolved := r.resolveLocked(ref)
	source := r.stops
	r.mu.RUnlock()
	if resolved != nil {
		return formula.Value{}, resolved
	}
	if source != nil {
		if stop, stopped := source.Stopped(library, name, version); stopped {
			return formula.Value{}, problem(codeVersionStopped, stoppedSentence(ref, stop))
		}
	}
	return release.Apply(args)
}

func (r *Registry) NoteOf(library string, name string, version int) (Note, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	release, found := r.release(library, name, version)
	if !found || version < 2 {
		return Note{}, false
	}
	return release.Note, true
}

func (r *Registry) MarkKnownWrong(library string, name string, version int, replacement int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ref := Ref{Library: library, Name: name, Version: version}
	h := r.functions[functionKey{library, name}]
	switch {
	case h == nil || version < 1 || version > len(h.releases):
		return fmt.Errorf("%w: %s does not exist", ErrBadMark, ref)
	case replacement <= version || replacement > len(h.releases):
		return fmt.Errorf("%w: the replacement %s must be a newer version that exists", ErrBadMark, Ref{Library: library, Name: name, Version: replacement})
	}
	if _, marked := h.marks[version]; marked {
		return fmt.Errorf("%w: %s is already marked", ErrBadMark, ref)
	}
	h.marks[version] = replacement
	return nil
}

func problem(code string, sentence string) *formula.Problem {
	return &formula.Problem{Code: code, Sentence: sentence}
}

func stoppedSentence(ref Ref, stop Stop) string {
	sentence := "`" + ref.String() + "` is stopped by the platform owner and cannot run."
	if reason := strings.TrimSpace(stop.Reason); reason != "" {
		sentence += " Reason: " + reason + "."
	}
	return sentence
}
