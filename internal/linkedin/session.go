package linkedin

import "strings"

// sanitizeSessionValue normalizes cookie values copied from browser DevTools or Vercel.
// Go rejects cookie values containing double quotes, so strip wrappers like "ajax:123".
func sanitizeSessionValue(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if i := strings.IndexByte(s, '='); i > 0 {
		name := strings.ToLower(s[:i])
		if name == "li_at" || name == "jsessionid" {
			s = strings.TrimSpace(s[i+1:])
		}
	}
	for {
		trimmed := strings.Trim(s, `"`)
		if trimmed == s {
			break
		}
		s = trimmed
	}
	return strings.TrimSpace(s)
}
