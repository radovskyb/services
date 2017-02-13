package user

import (
	"database/sql"
	"flag"
	"os"
	"testing"

	"github.com/radovskyb/services/user"
)

const (
	testEmail    = "radovskyb@gmail.com"
	testUsername = "radovskyb"

	// Constants used for testing with a real database.
	dsn              = "root:root@/golang"
	dropUserTableSQL = `DROP TABLE IF EXISTS users;`
)

func mysqlStoreSetup(t *testing.T) (UserStore, func()) {
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

	_, err = db.Exec(createUserTableSQL)
	if err != nil {
		t.Fatal(err)
	}

	// Insert a user into the database.
	_, err = db.Exec(
		"INSERT INTO users (email, username) VALUES (?, ?)",
		testEmail, testUsername,
	)
	if err != nil {
		t.Fatal(err)
	}
	return &mysqlStore{db}, teardown
}

func mockStoreSetup(t *testing.T) (UserStore, func()) {
	us := NewMockStore()
	// Insert a user into the database.
	u := &user.User{Email: testEmail, Username: testUsername}
	if err := us.Create(u); err != nil {
		t.Fatal(err)
	}
	teardown := func() {
		us = NewMockStore()
	}
	return us, teardown
}

var setupDB func(t *testing.T) (UserStore, func())

func TestMain(m *testing.M) {
	all := flag.Bool("all", false, "run all database implementations")
	flag.Parse()
	if !*all {
		setupDB = mockStoreSetup
		os.Exit(m.Run())
	}
	testCases := []struct {
		setup func(t *testing.T) (UserStore, func())
	}{
		{mockStoreSetup},
		{mysqlStoreSetup},
	}
	for _, tc := range testCases {
		setupDB = tc.setup
		code := m.Run()
		if code != 0 {
			os.Exit(code)
		}
	}
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
}

func TestCreateUser(t *testing.T) {
	us, teardown := setupDB(t)
	defer teardown()

	u := &user.User{
		Email:    testEmail,
		Username: testUsername,
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

// Split CreateUser into the following tests:
// func TestCantCreateWithEmptyFields(t *testing.T) {}
// func TestCantCreateWithDupEmail(t *testing.T) {}
// func TestCantCreateWithDupUsername(t *testing.T) {}
// func TestCreateUserSuccessfully(t *testing.T) {}

// func TestUpdateUser(t *testing.T) {} ...
// func TestDeleteUser(t *testing.T) {} ...
