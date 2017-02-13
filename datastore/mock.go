package user

import (
	"sync"

	"github.com/radovskyb/services/user"
)

type mockStore struct {
	mu    *sync.Mutex          // Protects the following.
	idCnt int64                // Auto incrementing id counter.
	users map[int64]*user.User // Id to User.

	// Mock user unique keys.
	emails    map[string]*user.User
	usernames map[string]*user.User
}

func NewMockStore() UserStore {
	return &mockStore{
		mu:        new(sync.Mutex),
		users:     make(map[int64]*user.User),
		emails:    make(map[string]*user.User),
		usernames: make(map[string]*user.User),
	}
}

func (s *mockStore) Create(u *user.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

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

func (s *mockStore) Get(id int64) (*user.User, error) {
	// Make sure the user exists.
	u, found := s.users[id]
	if !found {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *mockStore) GetByEmail(email string) (*user.User, error) {
	// Make sure the user exists.
	u, found := s.emails[email]
	if !found {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (s *mockStore) Update(u *user.User) error {
	// Check if the user exists.
	old, found := s.users[u.Id]
	if !found {
		return ErrUserNotFound
	}
	// Update the email's key.
	if old.Email != u.Email {
		// Make sure the new email doesn't already exist.
		if _, found := s.emails[u.Email]; found {
			return ErrDuplicateEmail
		}
		// Delete the old email.
		delete(s.emails, old.Email)
		// Add the new email.
		s.emails[u.Email] = u
	}
	// Update the username's key.
	if old.Username != u.Username {
		// Make sure the new username doesn't already exist.
		if _, found := s.usernames[u.Username]; found {
			return ErrDuplicateUsername
		}
		// Delete the old username.
		delete(s.usernames, old.Username)
		// Add the new username.
		s.usernames[u.Username] = u
	}
	// Update the user.
	s.users[u.Id] = u
	return nil
}

func (s *mockStore) Delete(id int64) error {
	u, found := s.users[id]
	if !found {
		return ErrUserNotFound
	}
	delete(s.users, id)
	delete(s.emails, u.Email)
	delete(s.usernames, u.Username)
	return nil
}
