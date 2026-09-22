// Package id generates the identifiers new records are given. It satisfies the
// application's IDGenerator port, so the use cases never call a random source
// directly and stay deterministic under test.
package id

import (
	"crypto/rand"
	"encoding/hex"
)

// NewUUID returns a random version 4 UUID in the canonical 8-4-4-4-12 form.
//
// Written out rather than pulled from a library because it is sixteen bytes
// from crypto/rand with six bits overwritten, and a dependency added for that
// is a dependency to keep patched. crypto/rand.Read never returns an error:
// since Go 1.24 it panics instead if the system source fails, which is the
// right outcome here anyway, as an identifier from a broken random source must
// never reach the database.
func NewUUID() string {
	raw := make([]byte, 16)
	rand.Read(raw) //nolint:errcheck // documented above: it cannot fail.

	raw[6] = (raw[6] & 0x0f) | 0x40 // version 4
	raw[8] = (raw[8] & 0x3f) | 0x80 // RFC 9562 variant

	encoded := make([]byte, 36)
	hex.Encode(encoded[0:8], raw[0:4])
	encoded[8] = '-'
	hex.Encode(encoded[9:13], raw[4:6])
	encoded[13] = '-'
	hex.Encode(encoded[14:18], raw[6:8])
	encoded[18] = '-'
	hex.Encode(encoded[19:23], raw[8:10])
	encoded[23] = '-'
	hex.Encode(encoded[24:36], raw[10:16])
	return string(encoded)
}
