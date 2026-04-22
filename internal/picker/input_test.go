package picker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseKey_Empty(t *testing.T) {
	ev, n := ParseKey([]byte{})
	assert.Equal(t, EventNone, ev)
	assert.Equal(t, 0, n)
}

func TestParseKey_Arrows(t *testing.T) {
	cases := []struct {
		name  string
		input []byte
		want  Event
	}{
		{"up", []byte{0x1b, '[', 'A'}, EventUp},
		{"down", []byte{0x1b, '[', 'B'}, EventDown},
		{"right", []byte{0x1b, '[', 'C'}, EventRight},
		{"left", []byte{0x1b, '[', 'D'}, EventLeft},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ev, n := ParseKey(tc.input)
			assert.Equal(t, tc.want, ev)
			assert.Equal(t, 3, n)
		})
	}
}

func TestParseKey_VimKeys(t *testing.T) {
	for b, want := range map[byte]Event{
		'k': EventUp, 'j': EventDown, 'h': EventLeft, 'l': EventRight,
	} {
		ev, n := ParseKey([]byte{b})
		assert.Equal(t, want, ev, "byte %q", b)
		assert.Equal(t, 1, n)
	}
}

func TestParseKey_ControlKeys(t *testing.T) {
	for b, want := range map[byte]Event{
		0x0d: EventEnter, // CR
		0x0a: EventEnter, // LF
		0x20: EventSpace,
		0x03: EventCtrlC,
		'q':  EventQ,
		'y':  EventY,
		'Y':  EventY,
		'n':  EventN,
		'N':  EventN,
		'e':  EventE,
		'E':  EventE,
		'?':  EventQuestion,
	} {
		ev, n := ParseKey([]byte{b})
		assert.Equal(t, want, ev, "byte 0x%02x", b)
		assert.Equal(t, 1, n)
	}
}

func TestParseKey_IncompleteEscape(t *testing.T) {
	ev, n := ParseKey([]byte{0x1b})
	assert.Equal(t, EventIncomplete, ev)
	assert.Equal(t, 0, n)
}

func TestParseKey_IncompleteCSI(t *testing.T) {
	ev, n := ParseKey([]byte{0x1b, '['})
	assert.Equal(t, EventIncomplete, ev)
	assert.Equal(t, 0, n)
}

func TestParseKey_BareEsc(t *testing.T) {
	ev, n := ParseKey([]byte{0x1b, 'x'})
	assert.Equal(t, EventEsc, ev)
	assert.Equal(t, 1, n)
}

func TestParseKey_UnknownCSI(t *testing.T) {
	ev, n := ParseKey([]byte{0x1b, '[', 'Z'})
	assert.Equal(t, EventNone, ev)
	assert.Equal(t, 3, n)
}

func TestParseKey_UnknownByte(t *testing.T) {
	ev, n := ParseKey([]byte{'x'})
	assert.Equal(t, EventNone, ev)
	assert.Equal(t, 1, n)
}
