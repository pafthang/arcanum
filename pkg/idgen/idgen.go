// Package idgen generates short random ids (15 chars alphanumeric).
package idgen

import (
	"crypto/rand"
	"encoding/binary"
)

const alphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// New returns a 15-character random id.
func New() string {
	var b [15]byte
	var buf [8]byte
	for i := 0; i < 15; i++ {
		if _, err := rand.Read(buf[:]); err != nil {
			// fallback: should never happen
			b[i] = alphabet[i%len(alphabet)]
			continue
		}
		n := binary.LittleEndian.Uint64(buf[:])
		b[i] = alphabet[n%uint64(len(alphabet))]
	}
	return string(b[:])
}
