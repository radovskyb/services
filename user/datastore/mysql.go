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
	res, err := s.db.Exec(
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
	// MySQL driver won't return an error for res.RowsAffected.
	affected, _ := res.RowsAffected()
	if affected != 1 {
		return ErrUserNotFound
	}
	return nil
}

func (s *mysqlRepo) Delete(id int64) error {
	res, err := s.db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		return err
	}
	// MySQL driver won't return an error for res.RowsAffected.
	affected, _ := res.RowsAffected()
	if affected != 1 {
		return ErrUserNotFound
	}
	return err
}

func (s *mysqlRepo) checkDupes(u *user.User) error {
	var id int64
	// Check if the email already exists.
	err1 := s.db.QueryRow(
		"SELECT id FROM users WHERE email = ?", u.Email,
	).Scan(&id)
	if id != 0 && id != u.Id {
		return ErrDuplicateEmail
	}
	// Check if the username already exists.
	err2 := s.db.QueryRow(
		"SELECT id FROM users WHERE username = ?", u.Username,
	).Scan(&id)
	if id != 0 && id != u.Id {
		return ErrDuplicateUsername
	}
	if err1 == nil {
		err1 = err2
	}
	return err1
}

func (s *mysqlRepo) Authenticate(email, pass string) (*user.User, error) {
	// Make sure the user exists.
	u := new(user.User)
	row := s.db.QueryRow("SELECT * FROM users WHERE email = ?", email)
	err := row.Scan(&u.Id, &u.Email, &u.Username, &u.Password)
	if err == sql.ErrNoRows {
		return nil, ErrUserNotFound
	}

	// Now match the user's password.
	if u.Password != pass {
		return nil, ErrWrongPassword
	}

	return u, nil
}
