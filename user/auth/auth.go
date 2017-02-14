package auth

import (
	"github.com/radovskyb/services/user"
	"github.com/radovskyb/services/user/datastore"
	"golang.org/x/crypto/bcrypt"
)

type Auth interface {
	// CreateUser hashes a user's password and then stores
	// the user in a user repository.
	CreateUser(u *user.User) error

	// AuthenticateUser authenticates a user from a user.
	//
	// Password is hashed and then compared to the user's
	// hashed password.
	AuthenticateUser(username, password string) error
}

// auth is the default implementation for Auth.
type auth struct {
	r datastore.UserRepository
}

// NewAuth creates a new Auth implementation for the specified
// user repository.
func NewAuth(userRepo datastore.UserRepository) Auth {
	return &auth{r: userRepo}
}

func (a *auth) CreateUser(u *user.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(u.Password), bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	u.Password = string(hashedPassword)
	return a.r.Create(u)
}

func (a *auth) AuthenticateUser(username, password string) error {
	u, err := a.r.GetByUsername(username)
	if err != nil {
		return err
	}
	return bcrypt.CompareHashAndPassword(
		[]byte(u.Password), []byte(password),
	)
}
