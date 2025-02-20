// Code generated by "go-enumerator"; DO NOT EDIT.

package app

import (
	"encoding"
	"fmt"
)

// String implements [fmt.Stringer]. If !s.Defined(), then a generated string is returned based on s's value.
func (s state) String() string {
	switch s {
	case invalid:
		return "invalid"
	case running:
		return "running"
	case stopping:
		return "stopping"
	case terminated:
		return "terminated"
	}
	return fmt.Sprintf("state(%d)", s)
}

// Bytes returns a byte-level representation of String(). If !s.Defined(), then a generated string is returned based on s's value.
func (s state) Bytes() []byte {
	switch s {
	case invalid:
		return []byte{'i', 'n', 'v', 'a', 'l', 'i', 'd'}
	case running:
		return []byte{'r', 'u', 'n', 'n', 'i', 'n', 'g'}
	case stopping:
		return []byte{'s', 't', 'o', 'p', 'p', 'i', 'n', 'g'}
	case terminated:
		return []byte{'t', 'e', 'r', 'm', 'i', 'n', 'a', 't', 'e', 'd'}
	}
	return []byte(fmt.Sprintf("state(%d)", s))
}

// Defined returns true if s holds a defined value.
func (s state) Defined() bool {
	switch s {
	case 0, 1, 2, 3:
		return true
	default:
		return false
	}
}

// Scan implements [fmt.Scanner]. Use [fmt.Scan] to parse strings into state values
func (s *state) Scan(scanState fmt.ScanState, verb rune) error {
	token, err := scanState.Token(true, nil)
	if err != nil {
		return err
	}

	switch string(token) {
	case "invalid":
		*s = invalid
	case "running":
		*s = running
	case "stopping":
		*s = stopping
	case "terminated":
		*s = terminated
	default:
		return fmt.Errorf("unknown state value: %s", token)
	}
	return nil
}

// Next returns the next defined state. If s is not defined, then Next returns the first defined value.
// Next() can be used to loop through all values of an enum.
//
//	s := state(0)
//	for {
//		fmt.Println(s)
//		s = s.Next()
//		if s == state(0) {
//			break
//		}
//	}
//
// The exact order that values are returned when looping should not be relied upon.
func (s state) Next() state {
	switch s {
	case invalid:
		return running
	case running:
		return stopping
	case stopping:
		return terminated
	case terminated:
		return invalid
	default:
		return invalid
	}
}

func _() {
	var x [1]struct{}
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the go-enumerator command to generate them again.
	_ = x[invalid-0]
	_ = x[running-1]
	_ = x[stopping-2]
	_ = x[terminated-3]
}

// MarshalText implements [encoding.TextMarshaler]
func (s state) MarshalText() ([]byte, error) {
	return s.Bytes(), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler]
func (s *state) UnmarshalText(x []byte) error {
	switch string(x) {
	case "invalid":
		*s = invalid
		return nil
	case "running":
		*s = running
		return nil
	case "stopping":
		*s = stopping
		return nil
	case "terminated":
		*s = terminated
		return nil
	default:
		return fmt.Errorf("failed to parse value %v into %T", x, *s)
	}
}

var (
	_ fmt.Stringer             = state(0)
	_ fmt.Scanner              = new(state)
	_ encoding.TextMarshaler   = state(0)
	_ encoding.TextUnmarshaler = new(state)
)
