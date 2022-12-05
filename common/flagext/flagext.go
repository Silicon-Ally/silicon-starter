// Package flagext provides shared helpers that implement the flag.Value
// interface.
package flagext

import "strings"

// StringList is a flag type that takes in a comma-separated list of values.
type StringList []string

func (sl *StringList) String() string {
	return strings.Join(*sl, ",")
}

func (sl *StringList) Set(in string) error {
	*sl = strings.Split(in, ",")
	return nil
}
