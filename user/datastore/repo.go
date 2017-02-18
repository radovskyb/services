package datastore

import (
	"errors"

	"github.com/radovskyb/services/user"
)

var (
	ErrDuplicateEmail    = errors.New("error: a user with that email already exists")
	ErrDuplicateUsername = errors.New("error: a user with that username already exists")
	ErrUserNotFound      = errors.New("error: user not found")
	ErrWrongPassword     = errors.New("error: incorrect password")
)

type UserRepository interface {
	Create(u *user.User) error
	Get(id int64) (*user.User, error)
	GetByEmail(email string) (*user.User, error)
	GetByUsername(username string) (*user.User, error)
	Update(u *user.User) error
	Delete(id int64) error
	Authenticate(email, password string) (*user.User, error)
}
