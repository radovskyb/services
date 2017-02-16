package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type Session interface {
	// LogInUser sets a user to logged in and stores their username
	// in the user's session.
	LogInUser(w http.ResponseWriter, r *http.Request, username string) error

	// LogOutUser removes a user from the user's session to log them out.
	LogOutUser(w http.ResponseWriter, r *http.Request) error

	// UserLoggedIn checks if any user is currently logged in.
	UserLoggedIn(r *http.Request) bool

	// CurrentUser returns the current logged in user's username.
	CurrentUser(r *http.Request) (string, error)
}

type session struct {
	cookiestore *sessions.CookieStore
}

func NewSession(keyPairs ...[]byte) Session {
	return &session{sessions.NewCookieStore(keyPairs...)}
}

func (s *session) LogInUser(w http.ResponseWriter, r *http.Request,
	username string) error {
	session, err := s.cookiestore.Get(r, "user_session")
	if err != nil {
		return err
	}
	session.Values["loggedin"] = true
	session.Values["username"] = username
	return session.Save(r, w)
}

func (s *session) LogOutUser(w http.ResponseWriter, r *http.Request) error {
	session, err := s.cookiestore.Get(r, "user_session")
	if err != nil {
		return err
	}
	if session.Values["loggedin"] != true {
		return nil
	}
	for key := range session.Values {
		delete(session.Values, key)
	}
	return session.Save(r, w)
}

func (s *session) UserLoggedIn(r *http.Request) bool {
	session, err := s.cookiestore.Get(r, "user_session")
	if err == nil && (session.Values["loggedin"] == true) {
		return true
	}
	return false
}

func (s *session) CurrentUser(r *http.Request) (string, error) {
	session, err := s.cookiestore.Get(r, "user_session")
	if err != nil {
		return "", err
	}
	return session.Values["username"].(string), nil
}
