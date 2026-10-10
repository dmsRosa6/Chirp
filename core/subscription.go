package core

import (
	"bytes"

	"github.com/dmsRosa6/Chirp/wire"
)

// Subscription is one SUB call: a pattern registered by a session under a
// session-scoped id.
type Subscription struct {
	ID      string
	Pattern string
	Session *Session
}

// subKey identifies a subscription inside the broker. Ids are only unique
// per session, so the session id is part of the key.
type subKey struct {
	session uint64
	id      string
}

func (s *Subscription) key() subKey {
	return subKey{session: s.Session.id, id: s.ID}
}

// deliver frames the message for this subscription and queues it on the
// session. It never blocks; a session that can't keep up is disconnected.
func (s *Subscription) deliver(msg Message) {
	var b bytes.Buffer
	wire.WriteMessage(&b, s.ID, msg.Subject, msg.Body) // writing to a Buffer can't fail
	s.Session.Push(b.Bytes())
}
