package auth

import (
	"testing"

	"github.com/radovskyb/services/user"
	"github.com/radovskyb/services/user/datastore"
)

const (
	testEmail    = "radovskyb@gmail.com"
	testUsername = "radovskyb"
	testPassword = "password123" // Only use bcrypt in production database.
)

func TestCreateUser(t *testing.T) {
	repo := datastore.NewMockRepo()
	auth := NewAuth(repo)

	u := &user.User{
		Email:    testEmail,
		Username: testUsername,
		Password: testPassword,
	}

	err := auth.CreateUser(u)
	if err != nil {
		t.Error(err)
	}

	// Get the inserted user from the database
	// and check that the password was hashed.
	u, err = repo.GetByUsername(testUsername)
	if err != nil {
		t.Error(err)
	}
	if u.Password == testPassword {
		t.Errorf("expected u.Password to be %s, got %s",
			testPassword, u.Password)
	}
}

func TestAuthenticateUser(t *testing.T) {
	repo := datastore.NewMockRepo()
	auth := NewAuth(repo)

	u := &user.User{
		Email:    testEmail,
		Username: testUsername,
		Password: testPassword,
	}

	err := auth.CreateUser(u)
	if err != nil {
		t.Error(err)
	}

	// Try to authenticate a user with correct credentials.
	err = auth.AuthenticateUser(testUsername, testPassword)
	if err != nil {
		t.Error(err)
	}

	// Try to authenticate a user that doesn't exist.
	err = auth.AuthenticateUser("invalidUsername", testPassword)
	if err != datastore.ErrUserNotFound {
		t.Error("expected err to be ErrUserNotFound, got %v", err)
	}

	// Try to authenticate a user with an invalid password.
	err = auth.AuthenticateUser(testUsername, "invalidPassword")
	if err == nil {
		t.Error("expected err to not be nil")
	}
}
