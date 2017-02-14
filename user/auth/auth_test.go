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

func TestValidateUser(t *testing.T) {
	auth := NewAuth(datastore.NewMockRepo())

	// Test empty fields.
	testCases := []struct {
		u *user.User
	}{
		{&user.User{}},
		{&user.User{Email: testEmail}},
		{&user.User{Email: testEmail, Username: testUsername}},
	}
	for _, tc := range testCases {
		err := auth.ValidateUser(tc.u)
		if err != ErrEmptyRequiredField {
			t.Errorf("expected err to be ErrEmptyRequiredField, got %v", err)
		}
	}

	// Test with all fields not blank.
	u := &user.User{
		Email:    testEmail,
		Username: testUsername,
		Password: testPassword,
	}
	err := auth.ValidateUser(u)
	if err == ErrEmptyRequiredField {
		t.Error("expected err to not be ErrEmptyRequiredField")
	}

	// Test invalid email.
	testCases = []struct {
		u *user.User
	}{
		{&user.User{
			Email: "a@a", Username: testUsername, Password: testPassword,
		}},
		{&user.User{
			Email: "aaa", Username: testUsername, Password: testPassword,
		}},
		{&user.User{
			Email: "a@a.c", Username: testUsername, Password: testPassword,
		}},
		{&user.User{
			Email: "%mail@$email.com", Username: testUsername, Password: testPassword,
		}},
	}
	for _, tc := range testCases {
		err = auth.ValidateUser(tc.u)
		if err != ErrInvalidEmail {
			t.Errorf("expected err to be ErrInvalidEmail, got %v", err)
		}
	}

	testCases = []struct {
		u *user.User
	}{
		{&user.User{
			Email: testEmail, Username: "a-b-c", Password: testPassword,
		}},
		{&user.User{
			Email: testEmail, Username: "%abc%", Password: testPassword,
		}},
		{&user.User{
			Email: testEmail, Username: "@123abc", Password: testPassword,
		}},
	}
	for _, tc := range testCases {
		err = auth.ValidateUser(tc.u)
		if err != ErrInvalidUsername {
			t.Errorf("expected err to be ErrInvalidUsername, got %v", err)
		}
	}

	// Test username length.
	testCases = []struct {
		u *user.User
	}{
		{&user.User{
			Email:    testEmail,
			Username: "a1",
			Password: testPassword,
		}},
		{&user.User{
			Email:    testEmail,
			Username: "abcdefghijklmnopqrstuvwxyz",
			Password: testPassword,
		}},
	}
	for _, tc := range testCases {
		err = auth.ValidateUser(tc.u)
		if err != ErrInvalidUsernameLength {
			t.Errorf("expected err to be ErrInvalidUsernameLength, got %v", err)
		}
	}

	// Test password length.
	u = &user.User{
		Email:    testEmail,
		Username: testUsername,
		Password: "12345",
	}
	err = auth.ValidateUser(u)
	if err != ErrPasswordTooShort {
		t.Errorf("expected err to be ErrInvalidPasswordLength, got %v", err)
	}
}

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

	// Try to create an invalid user.
	u.Password = ""
	err = auth.CreateUser(u)
	if err == nil {
		t.Errorf("expected there to be a validation error, got %v", err)
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
