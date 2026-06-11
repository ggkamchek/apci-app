package admin

import (
	"regexp"
	"strings"
)

// Canonical account IDs in admin UI always include hyphens.
// google/uuid.Parse also accepts 32 hex chars without hyphens, which
// collides with full device hashes — so lookup must not use Parse alone.
var canonicalAccountID = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

func isLookupAccountID(query string) bool {
	return canonicalAccountID.MatchString(strings.TrimSpace(query))
}
