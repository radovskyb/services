package session

import (
	"errors"
	"net/http"

	"github.com/gorilla/sessions"
)

var (
	ErrUserNotSet      = errors.New("user is not set for user_session")
	ErrUserNotLoggedIn = errors.New("user is not logged in")
)

type Session interface {
	// LogInUser sets a user to logged in and stores their username
	// in the user's session.
	LogInUser(w http.ResponseWriter, r *http.Request, username string) error

	// LogOutUser removes the current logged in user from the user's session.
	LogOutUser(w http.ResponseWriter, r *http.Request) error

	// UserLoggedIn checks if any user is currently logged in.
	UserLoggedIn(r *http.Request) bool

	// CurrentUser returns the current logged in user's username.
	CurrentUser(r *http.Request) (string, error)
}

// session is the default implementation for Session.
type session struct {
	cookiestore *sessions.CookieStore
}

func NewSession(store *sessions.CookieStore) Session {
	return &session{store}
}

func (s *session) LogInUser(w http.ResponseWriter, r *http.Request,
	username string) error {
	sess, err := s.cookiestore.Get(r, "user_session")
	if err != nil {
		return err
	}
	sess.Values["loggedin"] = true
	sess.Values["username"] = username
	return sess.Save(r, w)
}

func (s *session) LogOutUser(w http.ResponseWriter, r *http.Request) error {
	sess, err := s.cookiestore.Get(r, "user_session")
	if err != nil {
		return err
	}
	if sess.Values["loggedin"] != true {
		return ErrUserNotLoggedIn
	}
	for key := range sess.Values {
		delete(sess.Values, key)
	}
	return sess.Save(r, w)
}

func (s *session) UserLoggedIn(r *http.Request) bool {
	sess, err := s.cookiestore.Get(r, "user_session")
	if err == nil && (sess.Values["loggedin"] == true) {
		return true
	}
	return false
}

func (s *session) CurrentUser(r *http.Request) (string, error) {
	sess, err := s.cookiestore.Get(r, "user_session")
	if err != nil {
		return "", err
	}
	username, ok := sess.Values["username"]
	if !ok {
		return "", ErrUserNotSet
	}
	return username.(string), nil
}
