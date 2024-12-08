package repository

import (
	"fmt"
	"testing"

	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/user"
	"github.com/stretchr/testify/assert"
	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

const (
	id       = "1"
	username = "User"
	password = "password"
)

func TestNewUserPostgresRepo(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewUserPostgresRepo(db)

	assert.NotNil(t, repo)
	assert.Equal(t, db, repo.DB)
}

func TestGetUserByUsernameSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "password"})
	expect := []*user.User{
		{
			ID:       id,
			Username: username,
			Password: password,
		},
	}

	for _, user := range expect {
		rows = rows.AddRow(user.ID, user.Username, user.Password)
	}

	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE username = (.+);").
		WithArgs(username).
		WillReturnRows(rows)

	repo := &UserPostgresRepo{
		DB: db,
	}

	getUser, err := repo.GetUserByUsername(username)
	assert.NoError(t, err)
	assert.Equal(t, getUser, expect[0])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByUsernameErrorNoUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "password"})

	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE username = (.+);").
		WithArgs(username).
		WillReturnRows(rows)

	repo := &UserPostgresRepo{
		DB: db,
	}

	getUser, err := repo.GetUserByUsername(username)
	assert.Error(t, err, "postgres get user: user not found")
	assert.Nil(t, getUser)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserByUsernameError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE username = (.+);").
		WithArgs(username).
		WillReturnError(fmt.Errorf("scan error"))

	repo := &UserPostgresRepo{
		DB: db,
	}

	_, err = repo.GetUserByUsername(username)
	assert.Error(t, err, "postgres get user: scan error")
	assert.EqualError(t, err, "postgres get user: scan error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "password"})
	expect := []*user.User{
		{
			ID:       id,
			Username: username,
			Password: password,
		},
	}
	for _, user := range expect {
		rows = rows.AddRow(user.ID, user.Username, user.Password)
	}

	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE username = (.+);").
		WithArgs(username).
		WillReturnRows(rows)

	repo := &UserPostgresRepo{
		DB: db,
	}

	getUser, err := repo.Login(username, password)
	fmt.Println()
	assert.NoError(t, err)
	assert.Equal(t, getUser, expect[0])
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginInvalidPass(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "password"})

	expect := []*user.User{
		{
			ID:       id,
			Username: username,
			Password: password,
		},
	}
	for _, user := range expect {
		rows = rows.AddRow(user.ID, user.Username, user.Password)
	}

	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE username = (.+);").
		WithArgs(username).
		WillReturnRows(rows)

	repo := &UserPostgresRepo{
		DB: db,
	}

	_, err = repo.Login(username, "Invalid Password")
	fmt.Println()
	assert.EqualError(t, err, myerrors.ErrBadPass.Error())
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestLoginError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"id", "username", "password"})

	mock.
		ExpectQuery("SELECT id, username, password FROM users WHERE username = (.+);").
		WithArgs(username).
		WillReturnRows(rows)

	repo := &UserPostgresRepo{
		DB: db,
	}

	_, err = repo.Login(username, password)
	assert.EqualError(t, err, "postgres login user: postgres get user: user not found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsUserExists(t *testing.T) { //
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"exists"}).AddRow(true)

	mock.
		ExpectQuery(`^SELECT EXISTS\(SELECT 1 FROM users WHERE username = (.+)\);$`).
		WithArgs(username).
		WillReturnRows(rows)

	repo := &UserPostgresRepo{
		DB: db,
	}

	err = isUserExists(repo.DB, username)
	assert.ErrorIs(t, err, myerrors.ErrUserExist)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestIsUserNotExists(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"exists"})

	username := "Not existing user"

	mock.
		ExpectQuery(`^SELECT EXISTS\(SELECT 1 FROM users WHERE username = (.+)\);$`).
		WithArgs(username).
		WillReturnRows(rows)

	repo := &UserPostgresRepo{
		DB: db,
	}

	err = isUserExists(repo.DB, username)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupSuccess(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.
		ExpectQuery(`^SELECT EXISTS\(SELECT 1 FROM users WHERE username = (.+)\);$`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.
		ExpectExec(`INSERT INTO users \(id, username, password\) VALUES \((.+), (.+), (.+)\);`).
		WithArgs(sqlmock.AnyArg(), username, password).
		WillReturnResult(sqlmock.NewResult(1, 1))

	repo := &UserPostgresRepo{
		DB: db,
	}

	user, err := repo.Signup(username, password)
	assert.NoError(t, err)
	// assert.ErrorIs(t, err, myerrors.ErrUserExist)
	assert.Equal(t, user.Username, username)
	assert.Equal(t, user.Password, password)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupError(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.
		ExpectQuery(`^SELECT EXISTS\(SELECT 1 FROM users WHERE username = (.+)\);$`).
		WithArgs(username).
		WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

	mock.
		ExpectExec(`INSERT INTO users \(id, username, password\) VALUES \((.+), (.+), (.+)\);`).
		WithArgs(sqlmock.AnyArg(), username, password).
		WillReturnError(fmt.Errorf("create user error"))

	repo := &UserPostgresRepo{
		DB: db,
	}

	_, err = repo.Signup(username, password)
	assert.EqualError(t, err, "postgres signup user: create user error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSignupNoUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	mock.
		ExpectQuery(`^SELECT EXISTS\(SELECT 1 FROM users WHERE username = (.+)\);$`).
		WithArgs(username).
		WillReturnError(fmt.Errorf("user already exists"))

	repo := &UserPostgresRepo{
		DB: db,
	}

	_, err = repo.Signup(username, password)
	assert.EqualError(t, err, "postgres signup user: user already exists")
	assert.NoError(t, mock.ExpectationsWereMet())
}
