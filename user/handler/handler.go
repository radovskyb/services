package handler

import (
	"net/http"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/radovskyb/services/user"
	"github.com/radovskyb/services/user/auth"
	"github.com/radovskyb/services/user/datastore"
	"github.com/radovskyb/services/user/session"
)

type Handler struct {
	r datastore.UserRepository
	a auth.Auth
	s session.Session
}

func NewHandler(r datastore.UserRepository, s *sessions.CookieStore) *Handler {
	return &Handler{
		r: r,
		a: auth.NewAuth(r),
		s: session.NewSession(s),
	}
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var (
		email    = r.FormValue("email")
		username = r.FormValue("username")
		password = r.FormValue("password")
	)

	u := &user.User{
		Email:    email,
		Username: username,
		Password: password,
	}

	// Create the user in the user repository.
	if err := h.a.CreateUser(u); err != nil {
		if auth.IsValidationErr(err) {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var (
		id       = r.FormValue("id")
		email    = r.FormValue("email")
		username = r.FormValue("username")
		password = r.FormValue("password")
	)

	// TODO: Add in field validation here.

	// Get the current logged in user's username from the session.
	cur, err := h.s.CurrentUser(r)
	if err != nil {
		if err == session.ErrUserNotSet {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert id to an integer.
	uid, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the user for the associated uid.
	u, err := h.r.Get(int64(uid))
	if err != nil {
		if err == datastore.ErrUserNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Make sure that u's username matches cur. (invalid username for session)
	if cur != u.Username {
		http.Error(w, datastore.ErrUserNotFound.Error(), http.StatusNotFound)
		return
	}

	// Hash the updated password.
	hashedPassword, err := h.a.HashPassword(password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Finally update the user with the new fields.
	err = h.r.Update(&user.User{
		Id:       int64(uid),
		Email:    email,
		Username: username,
		Password: string(hashedPassword),
	})
	if err != nil {
		if err == datastore.ErrUserNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UserLogin(w http.ResponseWriter, r *http.Request) {
	var (
		email    = r.FormValue("email")
		password = r.FormValue("password")
	)

	// Make sure the email or password isn't empty.
	//
	// Full validation isn't required, since AuthenticateUser will
	// simply return an error, but it saves a datastore call if
	// at least a blank email or password check is in place.
	if email == "" || password == "" {
		http.Error(w, auth.ErrEmptyRequiredField.Error(), http.StatusBadRequest)
		return
	}

	// Authenticate the user.
	u, err := h.a.AuthenticateUser(email, password)
	if err != nil {
		switch err {
		case datastore.ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		case auth.ErrWrongPassword:
			http.Error(w, err.Error(), http.StatusUnauthorized)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Set the username to logged in for the session.
	err = h.s.LogInUser(w, r, u.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) UserLogout(w http.ResponseWriter, r *http.Request) {
	// Log out the currently logged in user.
	err := h.s.LogOutUser(w, r)
	if err != nil {
		if err == session.ErrUserNotLoggedIn {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
