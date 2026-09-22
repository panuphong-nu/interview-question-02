package token

import (
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"
)

// newTokenID returns a value unique to one issued token. Randomness alone
// would do; the timestamp prefix simply makes the identifiers sortable when
// they turn up in a log.
func newTokenID(issuedAt time.Time) string {
	suffix := make([]byte, 8)
	if _, err := rand.Read(suffix); err != nil {
		// crypto/rand does not fail on any supported platform, and Go 1.24
		// made Read panic rather than return an error there. Falling back to
		// the timestamp alone keeps sign in working if that ever changes.
		return strconv.FormatInt(issuedAt.UnixNano(), 36)
	}
	return strconv.FormatInt(issuedAt.Unix(), 36) + "-" +
		base64.RawURLEncoding.EncodeToString(suffix)
}
