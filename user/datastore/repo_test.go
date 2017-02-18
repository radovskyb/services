package datastore

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/radovskyb/services/user"
)

const (
	testEmail    = "radovskyb@gmail.com"
	testUsername = "radovskyb"
	testPassword = "password123" // Only use bcrypt in production database.

	// Constants used for testing with a real database.
	dsn              = "root:root@/golang"
	dropUserTableSQL = `DROP TABLE IF EXISTS users`
)

var setupDB func(t *testing.T) (UserRepository, func())

func TestMain(m *testing.M) {
	all := flag.Bool("all", false, "run all database implementations")
	flag.Parse()
	if !*all {
		setupDB = mockRepoSetup
		os.Exit(m.Run())
	}
	testCases := []struct {
		name  string
		setup func(t *testing.T) (UserRepository, func())
	}{
		{"mock", mockRepoSetup},
		{"mysql", mysqlRepoSetup},
	}
	for _, tc := range testCases {
		fmt.Println(tc.name)
		setupDB = tc.setup
		code := m.Run()
		if code != 0 {
			os.Exit(code)
		}
	}
}

func mysqlRepoSetup(t *testing.T) (UserRepository, func()) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatal(err)
	}

	teardown := func() {
		_, err := db.Exec(dropUserTableSQL)
		if err != nil {
			t.Error(err)
		}
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}

	_, err = db.Exec(dropUserTableSQL)
	if err != nil {
		t.Fatal(err)
	}

	us, err := NewMySQLRepo(db)
	if err != nil {
		t.Fatal(err)
	}

	// Insert a user into the database.
	_, err = db.Exec(
		"INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		testEmail, testUsername, testPassword,
	)
	if err != nil {
		t.Fatal(err)
	}

	return us, teardown
}

func mockRepoSetup(t *testing.T) (UserRepository, func()) {
	us := NewMockRepo()
	// Insert a user into the database.
	u := &user.User{
		Email:    testEmail,
		Username: testUsername,
		Password: testPassword,
	}
	if err := us.Create(u); err != nil {
		t.Fatal(err)
	}
	teardown := func() {
		mr, ok := us.(*mockRepo)
		if !ok {
			t.Fatal("error converting us to *mockRepo")
		}
		mr.Close()
	}
	return us, teardown
}

func TestGetUser(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	u, err := us.Get(1)
	if err != nil {
		t.Fatal(err)
	}

	if u.Email != testEmail {
		t.Errorf("expected email to be %s, got %s", testEmail, u.Email)
	}

	if u.Username != testUsername {
		t.Errorf("expected username to be %s, got %s", testUsername, u.Username)
	}

	// Try to get a user that doesn't exist.
	_, err = us.Get(2)
	if err != ErrUserNotFound {
		t.Error("expected error to be ErrUserNotFound")
	}
}

func TestGetByUsername(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	u1, err := us.GetByUsername(testUsername)
	if err != nil {
		t.Fatal(err)
	}

	u2, err := us.GetByUsername(u1.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Should be different pointers.
	if u1 == u2 {
		t.Error("expected u1 and u2 to be different pointers")
	}
}

func TestCreateUserAfterTeardown(t *testing.T) {
	us, teardown := setupDB(t)

	teardown()

	// Try to create a new user.
	err := us.Create(&user.User{})
	if err == nil {
		t.Error("expected err to not be nil")
	}
}

func TestCreateUser(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	u := &user.User{
		Email:    testEmail,
		Username: testUsername,
		Password: testPassword,
	}

	err := us.Create(u)
	if err != ErrDuplicateEmail {
		t.Error("expected email to already exist")
	}

	fixedEmail, fixedUsername := "example_user@gmail.com", "example_user"

	// Change the email.
	u.Email = fixedEmail

	err = us.Create(u)
	if err != ErrDuplicateUsername {
		t.Error("expected username to already exist")
	}

	// Change the username.
	u.Username = fixedUsername

	err = us.Create(u)
	if err != nil {
		t.Error("expected user to be created successfully")
	}

	// Make sure the user was created.
	u, err = us.Get(4)
	if err != nil {
		t.Error(err)
	}

	if u.Email != fixedEmail {
		t.Errorf("expected email to be %s, got %s", fixedEmail, u.Email)
	}

	if u.Username != fixedUsername {
		t.Errorf("expected username to be %s, got %s", fixedUsername, u.Username)
	}
}

func TestUpdateUser(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	var (
		newEmail    = "example_user@gmail.com"
		newUsername = "example_user"
		newPassword = "password456"
	)

	// Create a new user.
	u, err := us.GetByEmail(testEmail)
	if err != nil {
		t.Error(err)
	}

	// Update the email.
	u.Email = newEmail
	err = us.Update(u)
	if err != nil {
		t.Error(err)
	}

	_, err = us.GetByEmail(testEmail)
	if err != ErrUserNotFound {
		t.Errorf("expected not to find user for email %s", testEmail)
	}

	// Update the username.
	u.Username = newUsername
	err = us.Update(u)
	if err != nil {
		t.Error(err)
	}

	_, err = us.GetByUsername(testUsername)
	if err != ErrUserNotFound {
		t.Errorf("expected not to find user for username %s", testUsername)
	}

	// Update the password.
	u.Password = newPassword
	err = us.Update(u)
	if err != nil {
		t.Error(err)
	}

	// Try to update a user that doesn't exist.
	u.Id = 2
	err = us.Update(u)
	if err != ErrUserNotFound {
		t.Error("expected error to be ErrUserNotFound")
	}
}

func TestUpdateUserAfterTeardown(t *testing.T) {
	us, teardown := setupDB(t)

	u, err := us.GetByEmail(testEmail)
	if err != nil {
		t.Error(err)
	}

	teardown()

	// Try to update the user.
	err = us.Update(u)
	if err == nil {
		t.Error("expected err to not be nil")
	}
}

func TestUpdateUserWithDupEmail(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	var (
		email    = "example_user@gmail.com"
		username = "example_user"
		password = "password123"
	)

	u := &user.User{
		Email:    email,
		Username: username,
		Password: password,
	}

	// Create a new user.
	err := us.Create(u)
	if err != nil {
		t.Error(err)
	}

	// Get the user that's already in the database.
	u1, err := us.GetByEmail(testEmail)
	if err != nil {
		t.Error(err)
	}

	// Set the email to another email that already exists.
	u1.Email = email

	// Set the email to another email that already exists.
	err = us.Update(u1)
	if err != ErrDuplicateEmail {
		t.Errorf("expected err to be ErrDuplicateEmail, got %v", err)
	}
}

func TestUpdateUserWithDupUsername(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	var (
		email    = "example_user@gmail.com"
		username = "example_user"
		password = "password123"
	)

	u := &user.User{
		Email:    email,
		Username: username,
		Password: password,
	}

	// Create a new user.
	err := us.Create(u)
	if err != nil {
		t.Error(err)
	}

	// Get the user that's already in the database.
	u1, err := us.GetByEmail(testEmail)
	if err != nil {
		t.Error(err)
	}

	// Set the username to another username that already exists.
	u1.Username = username

	// Set the email to another email that already exists.
	err = us.Update(u1)
	if err != ErrDuplicateUsername {
		t.Errorf("expected err to be ErrDuplicateUsername, got %v", err)
	}
}

func TestDeleteUserAfterTeardown(t *testing.T) {
	us, teardown := setupDB(t)

	teardown()

	// Try to delete a user.
	err := us.Delete(1)
	if err == nil {
		t.Error("expected err to not be nil")
	}
}

func TestDeleteUser(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	// Try to delete a user that doesn't exist.
	err := us.Delete(2)
	if err != ErrUserNotFound {
		t.Error("expected error to be ErrUserNotFound")
	}

	// Delete the test user from the datastore.
	err = us.Delete(1)
	if err != nil {
		t.Error(err)
	}
}

func TestAuthenticateUser(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	// Try to authenticate a user with correct details.
	u, err := us.Authenticate(testEmail, testPassword)
	if err != nil {
		t.Error(err)
	}
	if u.Username != testUsername {
		t.Errorf("expected u.Username to be %s, got %s",
			testUsername, u.Username)
	}

	// Try to authenticate a user with an email that doesn't exist.
	u, err = us.Authenticate("notfound@example.com", testPassword)
	if err != ErrUserNotFound {
		t.Errorf("expected err to be ErrUserNotFound, got %v", err)
	}

	// Try to authenticate a user with an incorrect password.
	u, err = us.Authenticate(testEmail, "wrongpassword")
	if err != ErrWrongPassword {
		t.Errorf("expected err to be ErrWrongPassword, got %v", err)
	}
}

// Test *mysqlRepo.checkDupes method.
func TestCheckDupesMySQL(t *testing.T) {
	us, teardown := setupDB(t)

	// Only test for *mysqlRepo.
	mr, ok := us.(*mysqlRepo)
	if !ok {
		return
	}

	teardown()

	err := mr.checkDupes(&user.User{})
	if err == nil {
		t.Error("expected error to not be nil")
	}
}
