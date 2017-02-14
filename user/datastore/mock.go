package datastore

import (
	"errors"
	"sync"

	"github.com/radovskyb/services/user"
)

type mockRepo struct {
	mu    *sync.Mutex          // Protects the following.
	idCnt int64                // Auto incrementing id counter.
	users map[int64]*user.User // Id to User.

	// Mock user unique keys.
	emails    map[string]*user.User
	usernames map[string]*user.User
}

func NewMockRepo() UserRepository {
	return &mockRepo{
		mu:        new(sync.Mutex),
		users:     make(map[int64]*user.User),
		emails:    make(map[string]*user.User),
		usernames: make(map[string]*user.User),
	}
}

func (s *mockRepo) Close() {
	s.users = nil
	s.emails = nil
	s.usernames = nil
}

func (s *mockRepo) Create(u *user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.users == nil {
		return errors.New("database is closed")
	}

	s.idCnt++

	// Check if the username or email already exists.
	if _, found := s.emails[u.Email]; found {
		return ErrDuplicateEmail
	}
	if _, found := s.usernames[u.Username]; found {
		return ErrDuplicateUsername
	}

	// Make sure u's id is set to idCnt.
	u.Id = s.idCnt

	// Store the user.
	s.users[s.idCnt] = u

	// Store the unique user keys (email and username).
	s.emails[u.Email] = u
	s.usernames[u.Username] = u

	return nil
}

func (s *mockRepo) Get(id int64) (*user.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Make sure the user exists.
	u, found := s.users[id]
	if !found {
		return nil, ErrUserNotFound
	}

	// Return a different user pointer so fields being modified
	// doesn't directly update the database.
	new := &user.User{
		Id:       u.Id,
		Email:    u.Email,
		Username: u.Username,
		Password: u.Password,
	}
	return new, nil
}

func (s *mockRepo) GetByEmail(email string) (*user.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Make sure the user exists.
	u, found := s.emails[email]
	if !found {
		return nil, ErrUserNotFound
	}
	// Return a different user pointer so fields being modified
	// doesn't directly update the database.
	new := &user.User{
		Id:       u.Id,
		Email:    u.Email,
		Username: u.Username,
		Password: u.Password,
	}
	return new, nil
}

func (s *mockRepo) GetByUsername(username string) (*user.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Make sure the user exists.
	u, found := s.usernames[username]
	if !found {
		return nil, ErrUserNotFound
	}
	// Return a different user pointer so fields being modified
	// doesn't directly update the database.
	new := &user.User{
		Id:       u.Id,
		Email:    u.Email,
		Username: u.Username,
		Password: u.Password,
	}
	return new, nil
}

func (s *mockRepo) Update(u *user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if the user exists.
	old, found := s.users[u.Id]
	if !found {
		return ErrUserNotFound
	}
	// Update the email's key.
	//
	// Make sure the new email doesn't already exist.
	if u2, found := s.emails[u.Email]; found {
		if u2.Id != old.Id {
			return ErrDuplicateEmail
		}
	}
	// Update the username's key.
	//
	// Make sure the new username doesn't already exist.
	if u2, found := s.usernames[u.Username]; found {
		if u2.Id != old.Id {
			return ErrDuplicateUsername
		}
	}

	// Update the user.
	//
	// Replace u instead so pointers can't be directly modified
	// from previously returned users from the Get methods.
	updated := &user.User{
		Id:       u.Id,
		Email:    u.Email,
		Username: u.Username,
		Password: u.Password,
	}

	s.users[u.Id] = updated

	// Delete the old email.
	delete(s.emails, old.Email)
	// Add the new email.
	s.emails[u.Email] = updated

	// Delete the old username.
	delete(s.usernames, old.Username)
	// Add the new username.
	s.usernames[u.Username] = updated

	return nil
}

func (s *mockRepo) Delete(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, found := s.users[id]
	if !found {
		return ErrUserNotFound
	}
	delete(s.users, id)
	delete(s.emails, u.Email)
	delete(s.usernames, u.Username)

	return nil
}
