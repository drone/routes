package cookie

import (
	"net/http"
	"time"

	"github.com/drone/routes/exp/cookie/authcookie"
)

// Sign signs and timestamps a cookie so it cannot be forged.
func Sign(cookie *http.Cookie, secret string, expires time.Time) {
	val := SignStr(cookie.Value, secret, expires)
	cookie.Value = val
}

// SignStr signs and timestamps a string so it cannot be forged.
// 
// Normally used via Sign, but provided as a separate method for
// non-cookie uses. To decode a value not stored as a cookie use the
// DecodeStr function.
func SignStr(value, secret string, expires time.Time) string {
	return authcookie.New(value, expires, []byte(secret))
}

// DecodeStr returns the given signed cookie value if it validates,
// else returns an empty string.
func Decode(cookie *http.Cookie, secret string) string {
	return DecodeStr(cookie.Value, secret)
}

// DecodeStr returns the given signed value if it validates,
// else returns an empty string.
func DecodeStr(value, secret string) string {
	return authcookie.Login(value, []byte(secret))
}

// Clear deletes the cookie with the given name.
func Clear(w http.ResponseWriter, r *http.Request, name string) {
	cookie := http.Cookie{
		Name:   name,
		Value:  "deleted",
		Path:   "/",
		Domain: r.URL.Host,
		MaxAge: -1,
	}

	http.SetCookie(w, &cookie)
}





