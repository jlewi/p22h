package logging

const (
	// Debug indicates verbosity level with logr.
	// With logr the verbosity is additive
	// so log.V(1).Info() means log at verbosity = info + 1
	Debug = 1
)
