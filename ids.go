package manager

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

// babelforce entity ids are unhyphenated 32-char hex on the wire; the API does not match the
// hyphenated UUID form. The Rust SDK's generated path helper already rewrites hyphenated UUIDs to
// the canonical form — this request editor applies the same rule on every generated client, so a
// hyphenated id (e.g. copied from another system) round-trips identically in all three SDKs
// instead of 404ing. Non-UUID segments pass through untouched.
var hyphenatedUUID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

func normalizeUUIDPath(_ context.Context, req *http.Request) error {
	segments := strings.Split(req.URL.Path, "/")
	changed := false
	for i, s := range segments {
		if hyphenatedUUID.MatchString(s) {
			segments[i] = strings.ReplaceAll(s, "-", "")
			changed = true
		}
	}
	if changed {
		req.URL.Path = strings.Join(segments, "/")
		// Plain hex needs no escaping; drop any stale pre-encoded form of the old path.
		req.URL.RawPath = ""
	}
	return nil
}
