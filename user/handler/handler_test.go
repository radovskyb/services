package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/gorilla/sessions"
	"github.com/radovskyb/services/user/auth"
	"github.com/radovskyb/services/user/datastore"
	"github.com/radovskyb/services/user/session"
)

var server *httptest.Server

// setup sets up and returns a new user handler.
func setup() *Handler {
	mockRepo := datastore.NewMockRepo()
	cs := sessions.NewCookieStore([]byte("secret-session"))
	userHandler := NewHandler(mockRepo, cs)
	return userHandler
}

func TestMain(m *testing.M) {
	server = httptest.NewServer(nil)
	code := m.Run()
	server.Close()
	os.Exit(code)
}

func TestRegisterUser(t *testing.T) {
	uh := setup()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	pf := url.Values{}
	pf.Set("email", email)
	pf.Set("username", username)
	pf.Set("password", password)
	req.Form = pf

	rr := httptest.NewRecorder()

	uh.RegisterUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected code to be 200, got %d", rr.Code)
	}

	usr, err := uh.r.GetByEmail(email)
	if err != nil {
		t.Fatal(err)
	}
	if usr.Email != email {
		t.Errorf("expected email to be %s, got %s", email, usr.Email)
	}
	if usr.Username != username {
		t.Errorf("expected username to be %s, got %s", username, usr.Username)
	}
	if usr.Id != 1 {
		t.Errorf("expected id to be 1, got %d", usr.Id)
	}
}

func TestRegisterUserWithInvalidData(t *testing.T) {
	uh := setup()

	var (
		email    = "invalid@email"
		username = "radovskyb"
		password = "password123"
	)

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	pf := url.Values{}
	pf.Set("email", email)
	pf.Set("username", username)
	pf.Set("password", password)
	req.Form = pf

	rr := httptest.NewRecorder()

	uh.RegisterUser(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected code to be 400, got %d", rr.Code)
	}
}

func TestRegisterUserAfterRepoClose(t *testing.T) {
	uh := setup()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	pf := url.Values{}
	pf.Set("email", email)
	pf.Set("username", username)
	pf.Set("password", password)
	req.Form = pf

	rr := httptest.NewRecorder()

	closer, ok := uh.r.(io.Closer)
	if !ok {
		t.Fatal("repo doesn't implement an io.Closer")
	}
	closer.Close()

	uh.RegisterUser(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected code to be 500, got %d", rr.Code)
	}
}

func TestUserLogin(t *testing.T) {
	uh := setup()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	// Register a user.
	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	pf := url.Values{}
	pf.Set("email", email)
	pf.Set("username", username)
	pf.Set("password", password)
	req.Form = pf

	rr := httptest.NewRecorder()

	uh.RegisterUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected code to be 200, got %d", rr.Code)
	}

	// Check that there's no user set to logged in for the session.
	loggedIn := uh.s.UserLoggedIn(req)
	if loggedIn {
		t.Error("expected no user to be logged in")
	}

	// Log the user in.
	uh.UserLogin(rr, req)

	// Check that now there is a user set to logged in for the session.
	loggedIn = uh.s.UserLoggedIn(req)
	if !loggedIn {
		t.Error("expected user to be logged in")
	}

	// Get the username of the logged in user.
	cur, err := uh.s.CurrentUser(req)
	if err != nil {
		t.Error(err)
	}

	// Verify that the correct username is returned.
	if cur != username {
		t.Errorf("expected logged in username to be %s, got %s",
			username, cur)
	}
}

func TestUserLoginWithInvalidData(t *testing.T) {
	uh := setup()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	// Register a user.
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	pf := url.Values{}
	pf.Set("email", email)
	pf.Set("username", username)
	pf.Set("password", password)
	req.Form = pf

	uh.RegisterUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected code to be 200, got %d", rr.Code)
	}

	// Try to log the user in with empty values.
	rr = httptest.NewRecorder()

	pf = url.Values{}
	pf.Set("email", "")
	pf.Set("password", "")
	req.Form = pf

	uh.UserLogin(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected code to be 400, got %d", rr.Code)
	}

	body := rr.Body.String()
	if strings.TrimSpace(body) != auth.ErrEmptyRequiredField.Error() {
		t.Errorf("expected body to be a required field empty error, got %s", body)
	}

	// Try to log the user in with a non existing user.
	rr = httptest.NewRecorder()

	pf = url.Values{}
	pf.Set("email", "doesntexist@example.com")
	pf.Set("password", password)
	req.Form = pf

	uh.UserLogin(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected code to be 404, got %d", rr.Code)
	}

	body = rr.Body.String()
	if strings.TrimSpace(body) != datastore.ErrUserNotFound.Error() {
		t.Errorf("expected body to be a user not found error, got %s", body)
	}

	// Try to log the user in with an incorrect password.
	rr = httptest.NewRecorder()

	pf = url.Values{}
	pf.Set("email", email)
	pf.Set("password", "wrongpassword")
	req.Form = pf

	uh.UserLogin(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected code to be 401, got %d", rr.Code)
	}

	body = rr.Body.String()
	if strings.TrimSpace(body) != auth.ErrWrongPassword.Error() {
		t.Errorf("expected body to be an incorrect password error, got %s", body)
	}
}

func TestUserLoginAfterRepoClose(t *testing.T) {
	uh := setup()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	// Register a user.
	rr := httptest.NewRecorder()

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	pf := url.Values{}
	pf.Set("email", email)
	pf.Set("username", username)
	pf.Set("password", password)
	req.Form = pf

	uh.RegisterUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected code to be 200, got %d", rr.Code)
	}

	// Try to log the user in.
	rr = httptest.NewRecorder()

	pf = url.Values{}
	pf.Set("email", email)
	pf.Set("password", password)
	req.Form = pf

	// Close the user repository.
	closer, ok := uh.r.(io.Closer)
	if !ok {
		t.Fatal("repo doesn't implement an io.Closer")
	}
	closer.Close()

	uh.UserLogin(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("expected code to be 500, got %d", rr.Code)
	}
}

func TestUserLogout(t *testing.T) {
	uh := setup()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Error(err)
	}
	pf := url.Values{}
	pf.Set("email", email)
	pf.Set("username", username)
	pf.Set("password", password)
	req.Form = pf

	rr := httptest.NewRecorder()

	// Try to log out a user when no user is currently logged in.
	uh.UserLogout(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected code to be 404, got %d", rr.Code)
	}
	body := rr.Body.String()
	if strings.TrimSpace(body) != session.ErrUserNotLoggedIn.Error() {
		t.Errorf("expected body to be a user not logged in error, got %s", body)
	}

	// Reset the response recorder.
	rr = httptest.NewRecorder()

	// Register a user.
	uh.RegisterUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected code to be 200, got %d", rr.Code)
	}

	// Log the user in.
	uh.UserLogin(rr, req)

	// Check that now there is a user set to logged in for the session.
	loggedIn := uh.s.UserLoggedIn(req)
	if !loggedIn {
		t.Error("expected user to be logged in")
	}

	// Log the user out.
	uh.UserLogout(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected code to be 200, got %d", rr.Code)
	}

	// Check that there's no user set to logged in for the session.
	loggedIn = uh.s.UserLoggedIn(req)
	if loggedIn {
		t.Error("expected no user to be logged in")
	}
}
