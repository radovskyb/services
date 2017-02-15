package handler

import (
	"net/http"

	"github.com/radovskyb/services/user"
	"github.com/radovskyb/services/user/auth"
	"github.com/radovskyb/services/user/datastore"
)

type Handler struct {
	r datastore.UserRepository
	a auth.Auth
}

func NewHandler(r datastore.UserRepository) *Handler {
	return &Handler{r: r, a: auth.NewAuth(r)}
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
