package svcutil

import "strings"

// First returns the first non-empty string from vals (after trimming whitespace).
func First(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
