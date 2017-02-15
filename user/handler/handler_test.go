package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/gorilla/mux"
	"github.com/radovskyb/services/user/datastore"
)

var mockRepo = datastore.NewMockRepo()

func newRouter() *mux.Router {
	r := mux.NewRouter()
	userHandler := NewHandler(mockRepo)
	r.HandleFunc("/register", userHandler.RegisterUser)
	return r
}

func newTestServer() (*httptest.Server, func()) {
	s := httptest.NewServer(newRouter())
	teardown := func() {
		// Close the server.
		s.Close()
		// Reset the mockRepo.
		mockRepo = datastore.NewMockRepo()
	}
	return s, teardown
}

func TestRegisterUser(t *testing.T) {
	server, teardown := newTestServer()
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
