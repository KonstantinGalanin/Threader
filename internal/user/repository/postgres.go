package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/user"
	"github.com/google/uuid"
)

type UserPostgresRepo struct {
	DB *sql.DB
}

func NewUserPostgresRepo(db *sql.DB) *UserPostgresRepo {
	return &UserPostgresRepo{
		DB: db,
	}
}

func isUserExists(db *sql.DB, username string) error {
	var exists bool
	row := db.QueryRow(CheckExists, username)
	err := row.Scan(&exists)
	if exists {
		return myerrors.ErrUserExist
	}
	return err
}

func (u *UserPostgresRepo) Login(username, password string) (*user.User, error) {
	user, err := u.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("postgres login user: %w", err)
	}

	if user.Password != password {
		return nil, myerrors.ErrBadPass
	}

	return user, nil
}

func (u *UserPostgresRepo) Signup(username, password string) (*user.User, error) {
	err := isUserExists(u.DB, username)
	if err != nil {
		return nil, fmt.Errorf("postgres signup user: %w", err)
	}

	user := &user.User{
		ID:       uuid.New().String(),
		Password: password,
		Username: username,
	}

	_, err = u.DB.Exec(CreateUser, user.ID, user.Username, user.Password)
	if err != nil {
		return nil, fmt.Errorf("postgres signup user: %w", err)
	}
	return user, nil
}

func (u *UserPostgresRepo) GetUserByUsername(username string) (*user.User, error) {
	user := &user.User{}

	row := u.DB.QueryRow(GetUser, username)
	err := row.Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("postgres get user: %w", myerrors.ErrNoUser)
		}
		return nil, fmt.Errorf("postgres get user: %w", err)
	}

	return user, nil
}
