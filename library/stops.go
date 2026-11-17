package library

type Stop struct {
	Reason string
}

type StopSource interface {
	Stopped(library string, name string, version int) (Stop, bool)
}
