package db

import (
	"database/sql"
	"errors"
)

type User struct {
	ID           int
	Username     string
	PasswordHash string
}

func CreateUser(username, passwordHash string) error {
	stmt, err := DB.Prepare("INSERT INTO users(username, password_hash) VALUES($1, $2)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, passwordHash)
	return err
}

func GetUserByUsername(username string) (*User, error) {
	stmt, err := DB.Prepare("SELECT id, username, password_hash FROM users WHERE username = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(username)
	var user User
	err = row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &user, nil
}

func GetUserByID(id int) (*User, error) {
	stmt, err := DB.Prepare("SELECT id, username, password_hash FROM users WHERE id = $1")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	row := stmt.QueryRow(id)
	var user User
	err = row.Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}
