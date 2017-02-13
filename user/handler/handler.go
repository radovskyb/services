package handler

import (
	"net/http"

	"github.com/radovskyb/services/user"
	"github.com/radovskyb/services/user/datastore"
)

type Handler struct{ repo datastore.UserRepository }

func NewHandler(repo datastore.UserRepository) *Handler {
	return &Handler{repo: repo}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var (
		email    = r.FormValue("email")
		username = r.FormValue("username")
		password = r.FormValue("password")
	)

	// TODO: Put user validation here.

	// TODO: Hash password here.
	hashedPassword := password // Use regular pass for now.

	u := &user.User{Email: email, Username: username, Password: hashedPassword}

	// Create the user in the user repository.
	err := h.repo.Create(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
