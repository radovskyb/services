package user

import (
	"database/sql"
	"errors"

	"github.com/go-sql-driver/mysql"
	"github.com/radovskyb/services/user"
)

const createUserTableSQL = `CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTO_INCREMENT,
	email VARCHAR(255) UNIQUE NOT NULL,
	username VARCHAR(25) UNIQUE NOT NULL,
	password VARCHAR(72) NOT NULL
);`

type mysqlStore struct{ db *sql.DB }

func NewMySQLStore(db *sql.DB) (UserStore, error) {
	_, err := db.Exec(createUserTableSQL)
	return &mysqlStore{db}, err
}

func (s *mysqlStore) Create(u *user.User) error {
	_, err := s.db.Exec(
		"INSERT INTO users (email, username) VALUES (?, ?)",
		u.Email, u.Username,
	)
	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if !ok {
			return errors.New("error converting to mysql error")
		}
		if mysqlErr.Number == 1062 {
			var exists bool
			// Check if the email already exists.
			err := s.db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)", u.Email,
			).Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			if exists {
				return ErrDuplicateEmail
			}
			// Check if the username already exists.
			err = s.db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)", u.Username,
			).Scan(&exists)
			if err != nil && err != sql.ErrNoRows {
				return err
			}
			if exists {
				return ErrDuplicateUsername
			}
		}
	}
	return err
}

func (s *mysqlStore) Get(id int64) (*user.User, error) {
	u := new(user.User)
	row := s.db.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *mysqlStore) GetByEmail(email string) (*user.User, error) {
	u := new(user.User)
	row := s.db.QueryRow("SELECT * FROM users WHERE email = ?", email)
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *mysqlStore) GetByUsername(username string) (*user.User, error) {
	u := new(user.User)
	row := s.db.QueryRow("SELECT * FROM users WHERE username = ?", username)
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *mysqlStore) Update(u *user.User) error {
	_, err := s.db.Exec(
		"UPDATE users SET email = ?, username = ?, password = ? WHERE id = ?",
		u.Email, u.Username, u.Password, u.Id,
	)
	return err
}

func (s *mysqlStore) Delete(id int64) error {
	res, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	affecteded, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affecteded != 1 {
		return ErrUserNotFound
	}
	return nil
}
