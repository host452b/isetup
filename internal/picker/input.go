package picker

// Event represents a keyboard event parsed from raw stdin bytes.
type Event int

const (
	EventNone       Event = iota // Unrecognized or empty buffer; consume consumed bytes and keep reading.
	EventIncomplete              // Buffer is a prefix of an escape sequence; caller should read more bytes.
	EventUp
	EventDown
	EventLeft
	EventRight
	EventSpace
	EventEnter
	EventEsc
	EventCtrlC
	EventQ
	EventY
	EventN
	EventE
	EventQuestion
)

// ParseKey inspects the prefix of buf and returns the event it represents
// together with the number of bytes consumed. When the prefix is an unfinished
// escape sequence, it returns (EventIncomplete, 0) and the caller must read
// more bytes before calling again.
func ParseKey(buf []byte) (Event, int) {
	if len(buf) == 0 {
		return EventNone, 0
	}
	switch buf[0] {
	case 0x1b: // ESC
		if len(buf) == 1 {
			return EventIncomplete, 0
		}
		if buf[1] == '[' {
			if len(buf) < 3 {
				return EventIncomplete, 0
			}
			switch buf[2] {
			case 'A':
				return EventUp, 3
			case 'B':
				return EventDown, 3
			case 'C':
				return EventRight, 3
			case 'D':
				return EventLeft, 3
			}
			return EventNone, 3
		}
		return EventEsc, 1
	case 0x03:
		return EventCtrlC, 1
	case 0x0d, 0x0a:
		return EventEnter, 1
	case 0x20:
		return EventSpace, 1
	case 'k':
		return EventUp, 1
	case 'j':
		return EventDown, 1
	case 'h':
		return EventLeft, 1
	case 'l':
		return EventRight, 1
	case 'q':
		return EventQ, 1
	case 'y', 'Y':
		return EventY, 1
	case 'n', 'N':
		return EventN, 1
	case 'e', 'E':
		return EventE, 1
	case '?':
		return EventQuestion, 1
	}
	return EventNone, 1
}
