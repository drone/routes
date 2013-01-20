package user

import (
	"net/url"
	"github.com/drone/routes/exp/context"
)

// Key used to store the user in the session
const userKey = "_user"

// User represents a user of the application.
type User struct {
    Id    string // the unique permanent ID of the user.
	Name  string // the human-readable ID of the user.
	Email string
	Photo string

    FederatedIdentity string
    FederatedProvider string

	// additional, custom Attributes
	Attrs map[string]string
}

// Decode will create a user from a URL Query string.
func Decode(v string) *User {
	values, err := url.ParseQuery(v)
	if err != nil {
		return nil
	}

	attrs := map[string]string{}
	for key, _ := range values {
		attrs[key]=values.Get(key)
	}

	return &User {
		Id    : values.Get("id"),
		Name  : values.Get("name"),
		Email : values.Get("email"),
		Photo : values.Get("photo"),
		Attrs : attrs,
	}
}

// Encode will encode a user as a URL query string.
func (u *User) Encode() string {
	values := url.Values{}

	// add custom attributes
	if u.Attrs != nil {
		for key, val := range u.Attrs {
			values.Set(key, val)
		}
	}

	values.Set("id", u.Id)
	values.Set("name", u.Name)
	values.Set("email", u.Email)
	values.Set("photo", u.Photo)
	return values.Encode()
}

// Current returns the currently logged-in user, or nil if the user is not
// signed in.
func Current(c *context.Context) *User {
	v := c.Values.Get(userKey)
	if v == nil {
		return nil
	}

	u, ok := v.(*User)
	if !ok {
		return nil
	}

	return u
}

// Set sets the currently logged-in user. This is typically used by middleware
// that handles user authentication.
func Set(c *context.Context, u *User) {
	c.Values.Set(userKey, u)
}
