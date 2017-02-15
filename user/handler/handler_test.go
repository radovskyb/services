package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/radovskyb/services/user/datastore"
)

// setup returns a new httptest server, the user's datastore
// and a teardown function.
func setup() (*httptest.Server, datastore.UserRepository, func()) {
	mockRepo := datastore.NewMockRepo()

	r := mux.NewRouter()
	userHandler := NewHandler(mockRepo)
	r.HandleFunc("/register", userHandler.RegisterUser)

	s := httptest.NewServer(r)
	teardown := func() {
		// Close the server.
		s.Close()
	}

	return s, mockRepo, teardown
}

func TestRegisterUser(t *testing.T) {
	server, mockRepo, teardown := setup()
	defer teardown()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	u, err := url.Parse(server.URL + "/register")
	if err != nil {
		t.Fatal(err)
	}

	q := u.Query()
	q.Set("email", email)
	q.Set("username", username)
	q.Set("password", password)

	resp, err := http.PostForm(u.String(), q)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected response code to be 200, got %d", resp.StatusCode)
	}

	usr, err := mockRepo.GetByEmail(email)
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
	server, _, teardown := setup()
	defer teardown()

	var (
		email    = "invalid@email"
		username = "radovskyb"
		password = "password123"
	)

	u, err := url.Parse(server.URL + "/register")
	if err != nil {
		t.Fatal(err)
	}

	q := u.Query()
	q.Set("email", email)
	q.Set("username", username)
	q.Set("password", password)

	resp, err := http.PostForm(u.String(), q)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected response code to be 400, got %d", resp.StatusCode)
	}
}

func TestRegisterUserAfterTeardown(t *testing.T) {
	server, mockRepo, teardown := setup()
	defer teardown()

	var (
		email    = "radovskyb@gmail.com"
		username = "radovskyb"
		password = "password123"
	)

	u, err := url.Parse(server.URL + "/register")
	if err != nil {
		t.Fatal(err)
	}

	q := u.Query()
	q.Set("email", email)
	q.Set("username", username)
	q.Set("password", password)

	closer, ok := mockRepo.(io.Closer)
	if !ok {
		t.Fatal("repo doesn't implement an io.Closer")
	}
	closer.Close()

	resp, err := http.PostForm(u.String(), q)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected response code to be 500, got %d", resp.StatusCode)
	}
}
