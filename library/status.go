package library

type Status struct {
	Ref
	Exists      bool
	Latest      int
	Newer       bool
	Note        Note
	KnownWrong  bool
	Replacement int
	Stopped     bool
	StopReason  string
}

func (r *Registry) Status(library string, name string, version int) Status {
	r.mu.RLock()
	status := Status{Ref: Ref{Library: library, Name: name, Version: version}, Latest: r.latest(library, name)}
	_, status.Exists = r.release(library, name, version)
	if status.Exists {
		status.Newer = status.Latest > version
		status.Replacement, status.KnownWrong = r.markOf(library, name, version)
		status.Note = r.noticeNote(library, name, status)
	}
	source := r.stops
	r.mu.RUnlock()
	if status.Exists && source != nil {
		if stop, stopped := source.Stopped(library, name, version); stopped {
			status.Stopped = true
			status.StopReason = stop.Reason
		}
	}
	return status
}

func (r *Registry) markOf(library string, name string, version int) (int, bool) {
	h := r.functions[functionKey{library, name}]
	if h == nil {
		return 0, false
	}
	replacement, marked := h.marks[version]
	return replacement, marked
}

func (r *Registry) noticeNote(library string, name string, status Status) Note {
	target := 0
	switch {
	case status.KnownWrong:
		target = status.Replacement
	case status.Newer:
		target = status.Latest
	}
	if target < 2 {
		return Note{}
	}
	release, _ := r.release(library, name, target)
	return release.Note
}
