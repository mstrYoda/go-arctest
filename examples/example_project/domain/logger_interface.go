package domain

// Logger defines the interface that should be used by services
// instead of concrete logger implementations
type Logger interface {
	Log(message string)
	LogError(err error)
}
