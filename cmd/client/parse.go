package main

import "strings"

// parseLine turns a line typed at the prompt into command arguments.
//
// Everything is split on whitespace except PUB, whose payload is the rest of
// the line verbatim. The command and subject are skipped by position, so a
// subject that happens to appear inside the command name ("PUB P hi") or the
// payload can't confuse the split.
func parseLine(line string) []string {
	cmd, rest := nextToken(line)
	if cmd == "" {
		return nil
	}
	if !strings.EqualFold(cmd, "PUB") {
		return strings.Fields(line)
	}

	subject, rest := nextToken(rest)
	if subject == "" {
		return []string{cmd}
	}
	payload := strings.TrimLeft(rest, " \t")
	if payload == "" {
		return []string{cmd, subject}
	}
	return []string{cmd, subject, payload}
}

// nextToken returns the first whitespace-delimited token of s and what
// follows it (including the separating whitespace).
func nextToken(s string) (token, rest string) {
	s = strings.TrimLeft(s, " \t")
	i := strings.IndexAny(s, " \t")
	if i < 0 {
		return s, ""
	}
	return s[:i], s[i:]
}
