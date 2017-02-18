package session

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/sessions"
)

var server *httptest.Server

const testUsername = "radovskyb"

func setup() Session {
	return NewSession(
		sessions.NewCookieStore([]byte("secret-session")),
	)
}

func TestMain(m *testing.M) {
	server = httptest.NewServer(nil)
	code := m.Run()
	server.Close()
	os.Exit(code)
}

func TestCurrentUserWithoutLogin(t *testing.T) {
	sess := setup()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Error(err)
	}

	username, err := sess.CurrentUser(req)
	if err != ErrUserNotSet {
		t.Errorf("expected err to be ErrUserNotSet, got %v", err)
	}

	if username != "" {
		t.Errorf("expected username to be blank, got %s", username)
	}
}

func TestLogInUser(t *testing.T) {
	sess := setup()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()

	// Log in the user.
	err = sess.LogInUser(rr, req, testUsername)
	if err != nil {
		t.Error(err)
	}

	// Get the username from the session.
	username, err := sess.CurrentUser(req)
	if err != nil {
		t.Error(err)
	}

	if username != testUsername {
		t.Errorf("expected username to be %s, got %s",
			testUsername, username)
	}
}

func TestLogOutUser(t *testing.T) {
	sess := setup()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Error(err)
	}

	rr := httptest.NewRecorder()

	// Log in the user.
	err = sess.LogInUser(rr, req, testUsername)
	if err != nil {
		t.Error(err)
	}

	// Get the username from the session.
	username, err := sess.CurrentUser(req)
	if err != nil {
		t.Error(err)
	}

	if username != testUsername {
		t.Errorf("expected username to be %s, got %s",
			testUsername, username)
	}

	// Log out the logged in user.
	err = sess.LogOutUser(rr, req)
	if err != nil {
		t.Error(err)
	}

	// Try to get a username from the session.
	username, err = sess.CurrentUser(req)
	if err != ErrUserNotSet {
		t.Errorf("expected err to be ErrUserNotSet, got %v", err)
	}

	// The username should now be empty.
	if username != "" {
		t.Errorf("expected username to be blank, got %s", username)
	}

	// Try to log out a user when no user is logged in.
	// Log out the logged in user.
	err = sess.LogOutUser(rr, req)
	if err != ErrUserNotLoggedIn {
		t.Errorf("expected err to be ErrUserNotLoggedIn, got %v", err)
	}
}

func TestUserLoggedIn(t *testing.T) {
	sess := setup()

	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Error(err)
	}

	// Check if the user is logged in for req.
	loggedIn := sess.UserLoggedIn(req)
	if loggedIn {
		t.Error("expected no user to be logged in")
	}

	rr := httptest.NewRecorder()

	// Log in the user.
	err = sess.LogInUser(rr, req, testUsername)
	if err != nil {
		t.Error(err)
	}

	// Check if the user is logged in for req.
	loggedIn = sess.UserLoggedIn(req)
	if !loggedIn {
		t.Error("expected user to be logged in")
	}
}
