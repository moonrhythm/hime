package hime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"
)

// ErrInvalidCookie is returned by a CookieSigner when a signed cookie value is
// malformed or fails verification.
var ErrInvalidCookie = errors.New("hime: invalid signed cookie")

// CookieSigner signs and verifies cookie values so tampering can be detected.
//
// Implementations must bind the cookie name into the signature (so a value
// signed for one cookie can not be reused under a different name) and must use
// a constant-time comparison when verifying.
type CookieSigner interface {
	// Sign returns the signed, encoded value to store in the cookie named name.
	Sign(name, value string) (string, error)

	// Verify checks signedValue for the cookie named name and returns the
	// original value, or an error if it is malformed or fails verification.
	Verify(name, signedValue string) (string, error)
}

// AddSignedCookie signs value with the app's CookieSigner and writes it as a
// cookie. It is the signed counterpart of AddCookie; opts is applied the same
// way. It panics if the app has no CookieSigner configured.
func (ctx *Context) AddSignedCookie(name, value string, opts *CookieOptions) error {
	signer := ctx.app.CookieSigner
	if signer == nil {
		panicf("no cookie signer configured")
	}
	signed, err := signer.Sign(name, value)
	if err != nil {
		return err
	}
	ctx.AddCookie(name, signed, opts)
	return nil
}

// SignedCookieValue verifies the named cookie with the app's CookieSigner and
// returns its value, or an empty string if the cookie is absent OR fails
// verification (tampered, wrong key, or wrong name). The two cases are
// intentionally indistinguishable — the safe default is to not trust the value
// either way. If you need to tell them apart (for example to log tampering),
// read the raw cookie with CookieValue and call the signer's Verify yourself.
// It panics if the app has no CookieSigner configured.
func (ctx *Context) SignedCookieValue(name string) string {
	signer := ctx.app.CookieSigner
	if signer == nil {
		panicf("no cookie signer configured")
	}
	signed := ctx.CookieValue(name)
	if signed == "" {
		return ""
	}
	value, err := signer.Verify(name, signed)
	if err != nil {
		return ""
	}
	return value
}

// HMACCookieSigner is a CookieSigner backed by HMAC-SHA256. The cookie name is
// bound into the signature and verification is constant time.
//
// It signs but does not encrypt: the value stays readable by the client. Serve
// signed cookies as Secure + HttpOnly over HTTPS to protect them in transit,
// and rely on cookie MaxAge for expiry (the signature itself never expires).
//
// Because the signature never expires on its own, rotating the key invalidates
// every existing cookie. To rotate without logging users out, keep verifying
// against the old key (try each signer in turn) until the old cookies' MaxAge
// has elapsed.
type HMACCookieSigner struct {
	key []byte
}

// NewHMACCookieSigner returns an HMACCookieSigner using key (32+ random bytes
// recommended). The key is copied, so the caller may reuse or zero its slice.
// It panics if key is empty, since an empty key provides no protection.
func NewHMACCookieSigner(key []byte) *HMACCookieSigner {
	if len(key) == 0 {
		panicf("cookie signer key must not be empty")
	}
	return &HMACCookieSigner{key: append([]byte(nil), key...)}
}

// mac computes HMAC-SHA256 over name, a NUL separator, then value. A cookie
// name can not contain NUL, so this uniquely encodes the (name, value) pair and
// binds the name into the signature.
func (s *HMACCookieSigner) mac(name string, value []byte) []byte {
	h := hmac.New(sha256.New, s.key)
	h.Write([]byte(name))
	h.Write([]byte{0})
	h.Write(value)
	return h.Sum(nil)
}

// Sign implements CookieSigner. The wire format is
// base64url(value) "." base64url(mac).
func (s *HMACCookieSigner) Sign(name, value string) (string, error) {
	mac := s.mac(name, []byte(value))
	return base64.RawURLEncoding.EncodeToString([]byte(value)) + "." +
		base64.RawURLEncoding.EncodeToString(mac), nil
}

// Verify implements CookieSigner.
func (s *HMACCookieSigner) Verify(name, signedValue string) (string, error) {
	encValue, encMAC, ok := strings.Cut(signedValue, ".")
	if !ok {
		return "", ErrInvalidCookie
	}
	value, err := base64.RawURLEncoding.DecodeString(encValue)
	if err != nil {
		return "", ErrInvalidCookie
	}
	gotMAC, err := base64.RawURLEncoding.DecodeString(encMAC)
	if err != nil {
		return "", ErrInvalidCookie
	}
	if !hmac.Equal(gotMAC, s.mac(name, value)) {
		return "", ErrInvalidCookie
	}
	return string(value), nil
}
