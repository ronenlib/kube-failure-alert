package notifier

// Payload holds details of the event that needs to be notified
type Payload struct {
	Kind string
	Namespace string
	Name string
	Error string
}

// Notifier send notification to the client
type Notifier interface {
	Notify(payload Payload) error 
}