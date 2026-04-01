package executor

import "regexp"

// sudoRe matches "sudo " as a standalone command prefix — at the start of a
// line, after a pipe, after && / || / ;, or after an opening parenthesis.
var sudoRe = regexp.MustCompile(`\bsudo\s+`)

// StripSudo removes all "sudo " invocations from a command string.
// Used when running as root (UID 0) where sudo is unnecessary and often absent.
func StripSudo(cmd string) string {
	return sudoRe.ReplaceAllString(cmd, "")
}
