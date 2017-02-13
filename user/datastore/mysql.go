package datastore

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"github.com/radovskyb/services/user"
)

const createUserTableSQL = `CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTO_INCREMENT,
	email VARCHAR(255) UNIQUE NOT NULL,
	username VARCHAR(25) UNIQUE NOT NULL,
	password VARCHAR(72) NOT NULL
);`

type mysqlRepo struct{ db *sql.DB }

func NewMySQLRepo(db *sql.DB) (UserRepository, error) {
	_, err := db.Exec(createUserTableSQL)
	return &mysqlRepo{db}, err
}

func (s *mysqlRepo) Create(u *user.User) error {
	_, err := s.db.Exec(
		"INSERT INTO users (email, username, password) VALUES (?, ?, ?)",
		u.Email, u.Username, u.Password,
	)
	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			if dupeErr := s.checkDupes(u); dupeErr != nil {
				return dupeErr
			}
		}
		if !ok {
			return fmt.Errorf("error converting to mysql error: %s", err.Error())
		}
	}
	return nil
}

func (s *mysqlRepo) Get(id int64) (*user.User, error) {
	u := new(user.User)
	row := s.db.QueryRow("SELECT * FROM users WHERE id = ?", id)
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *mysqlRepo) GetByEmail(email string) (*user.User, error) {
	u := new(user.User)
	row := s.db.QueryRow("SELECT * FROM users WHERE email = ?", email)
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *mysqlRepo) GetByUsername(username string) (*user.User, error) {
	u := new(user.User)
	row := s.db.QueryRow("SELECT * FROM users WHERE username = ?", username)
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}
	return u, err
}

func (s *mysqlRepo) Update(u *user.User) error {
	_, err := s.db.Exec(
		"UPDATE users SET email = ?, username = ?, password = ? WHERE id = ?",
		u.Email, u.Username, u.Password, u.Id,
	)
	if err != nil {
		mysqlErr, ok := err.(*mysql.MySQLError)
		if ok && mysqlErr.Number == 1062 {
			if dupeErr := s.checkDupes(u); dupeErr != nil {
				return dupeErr
			}
		}
		if !ok {
			return fmt.Errorf("error converting to mysql error: %s", err.Error())
		}
	}
	return nil
}

func (s *mysqlRepo) Delete(id int64) error {
	res, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected != 1 {
		return ErrUserNotFound
	}
	return nil
}

func (s *mysqlRepo) checkDupes(u *user.User) error {
	var id int64
	// Check if the email already exists.
	err := s.db.QueryRow(
		"SELECT id FROM users WHERE email = ?", u.Email,
	).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if id != u.Id {
		return ErrDuplicateEmail
	}
	// Check if the username already exists.
	err = s.db.QueryRow(
		"SELECT id FROM users WHERE username = ?", u.Username,
	).Scan(&id)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if id != u.Id {
		return ErrDuplicateUsername
	}
	return nil
}
