package hime

import (
	"encoding/base64"
	"encoding/json"
)

// flashCookieName is the cookie used to carry flash messages across a redirect.
const flashCookieName = "flash"

// AddFlash queues a flash message under category, to be read exactly once on a
// later request via Flashes. Messages are stored in a cookie, so they survive
// the redirect of the post/redirect/get pattern. Calls accumulate.
//
// Flash messages are not encrypted or signed; do not put secrets in them.
func (ctx *Context) AddFlash(category, value string) {
	if ctx.flash == nil {
		ctx.flash = map[string][]string{}
	}
	ctx.flash[category] = append(ctx.flash[category], value)

	b, _ := json.Marshal(ctx.flash)
	ctx.AddCookie(flashCookieName, base64.RawURLEncoding.EncodeToString(b), &CookieOptions{
		Path:     "/",
		HttpOnly: true,
	})
}

// Flashes returns the flash messages queued on a previous request, keyed by
// category, and clears them so the next request will not see them again. It
// returns nil when there are none.
//
// Call Flashes before writing the response: the clear is sent as a Set-Cookie
// header, which has no effect once the response headers have been written.
func (ctx *Context) Flashes() map[string][]string {
	v := ctx.CookieValue(flashCookieName)
	if v == "" {
		return nil
	}

	// Clear the cookie regardless of whether it decodes, so a corrupt value
	// can not get stuck.
	ctx.DelCookie(flashCookieName, &CookieOptions{Path: "/"})

	b, err := base64.RawURLEncoding.DecodeString(v)
	if err != nil {
		return nil
	}
	var m map[string][]string
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
