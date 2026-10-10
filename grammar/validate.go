package grammar

import (
	"errors"
	"fmt"
	"strings"
)

const (
	MaxSubjectLen = 255
	MaxSubIDLen   = 64
)

// ValidateSubject checks a concrete subject (what PUB takes): dot-separated
// tokens, no empty tokens, no wildcard.
func ValidateSubject(s string) error { return validate(s, false) }

// ValidatePattern checks what SUB takes: a subject, except that the last
// token may end in the wildcard.
func ValidatePattern(s string) error { return validate(s, true) }

// ValidateSubID checks a subscription id chosen by the client.
func ValidateSubID(id string) error {
	if id == "" {
		return errors.New("sub id is empty")
	}
	if len(id) > MaxSubIDLen {
		return fmt.Errorf("sub id too long (max %d bytes)", MaxSubIDLen)
	}
	return nil
}

func validate(s string, allowWildcard bool) error {
	if s == "" {
		return errors.New("subject is empty")
	}
	if len(s) > MaxSubjectLen {
		return fmt.Errorf("subject too long (max %d bytes)", MaxSubjectLen)
	}

	tokens := strings.Split(s, SubjectDelimiter)
	for i, tok := range tokens {
		if tok == "" {
			return errors.New("subject has an empty token")
		}
		body := tok
		if strings.HasSuffix(tok, SubjectWildcard) {
			if !allowWildcard {
				return errors.New("wildcard not allowed here")
			}
			if i != len(tokens)-1 {
				return errors.New("wildcard is only allowed in the last token")
			}
			body = strings.TrimSuffix(tok, SubjectWildcard)
		}
		for _, r := range body {
			if !allowedRune(r) {
				return fmt.Errorf("invalid character %q in subject", r)
			}
		}
	}
	return nil
}

func allowedRune(r rune) bool {
	return r >= 'a' && r <= 'z' ||
		r >= 'A' && r <= 'Z' ||
		r >= '0' && r <= '9' ||
		r == '_' || r == '-'
}
