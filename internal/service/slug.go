package service

import (
	"crypto/rand"
	"math/big"
	"strings"
)

// slugify reduces a title to a URL-safe lowercase token.
func slugify(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	lastHyphen := false
	for _, r := range s {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			lastHyphen = false
		case !lastHyphen:
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	out := strings.Trim(b.String(), "-")
	if len(out) > 60 {
		out = strings.Trim(out[:60], "-")
	}
	if out == "" {
		out = "form"
	}
	return out
}

const slugAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// randSuffix returns a short random base-36 token for slug uniqueness.
func randSuffix(n int) string {
	b := make([]byte, n)
	for i := range b {
		idx, _ := rand.Int(rand.Reader, big.NewInt(int64(len(slugAlphabet))))
		b[i] = slugAlphabet[idx.Int64()]
	}
	return string(b)
}
