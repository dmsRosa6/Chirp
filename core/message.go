package core

// Message is one publish, as seen by the broker.
type Message struct {
	Subject string
	Body    string
	Time    int64 // unix nanoseconds, set by the broker on publish
}
