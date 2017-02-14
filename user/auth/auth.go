package auth

import (
	"errors"
	"regexp"
	"unicode"

	"github.com/radovskyb/services/user"
	"github.com/radovskyb/services/user/datastore"
	"golang.org/x/crypto/bcrypt"
)

var emailRegexp = regexp.MustCompile("[A-Z0-9a-z._%+-]+@[A-Za-z0-9.-]+\\.[A-Za-z]{2,6}")

var (
	ErrEmptyRequiredField    = errors.New("error: required user field is empty")
	ErrInvalidUsernameLength = errors.New("error: username must be between 3 - 25 characters")
	ErrInvalidEmail          = errors.New("error: email is invalid")
	ErrPasswordTooShort      = errors.New("error: password is too short (must be at least 6 characters)")
	ErrInvalidUsername       = errors.New("error: username is invalid (can only contain numbers and letters)")
)

type Auth interface {
	// CreateUser hashes a user's password and then stores
	// the user in a user repository.
	CreateUser(u *user.User) error

	// ValidateUser checks to see if the fields of a user are
	// valid to be used with the user's repository.
	ValidateUser(u *user.User) error

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

func (a *auth) ValidateUser(u *user.User) error {
	if u.Email == "" || u.Username == "" || u.Password == "" {
		return ErrEmptyRequiredField
	}
	if !emailRegexp.MatchString(u.Email) {
		return ErrInvalidEmail
	}
	if !isAlphanumeric(u.Username) {
		return ErrInvalidUsername
	}
	if len(u.Username) < 3 || len(u.Username) > 25 {
		return ErrInvalidUsernameLength
	}
	if len(u.Password) < 6 {
		return ErrPasswordTooShort
	}
	return nil
}

// isAlphanumeric checks whether a string contains only alphanumeric
// unicode characters.
func isAlphanumeric(str string) bool {
	for _, r := range str {
		if !(unicode.IsLetter(r) || unicode.IsNumber(r)) {
			return false
		}
	}
	return true
}

func (a *auth) CreateUser(u *user.User) error {
	if err := a.ValidateUser(u); err != nil {
		return err
	}
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
