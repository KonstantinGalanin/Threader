package myerrors

import "errors"

var (
	ErrBadPass           = errors.New("invalid password")
	ErrNoPost            = errors.New("no post with this id")
	ErrNoUser            = errors.New("user not found")
	ErrUserExist         = errors.New("user with this username already exist")
	ErrInvalidChars      = errors.New("contains invalid characters")
	ErrNeedMoreChars     = errors.New("must be more than 8 characters")
	ErrEmptyPostID       = errors.New("empty id in url")
	ErrEmptyCategory     = errors.New("empty category in url")
	ErrEmptyCommentID    = errors.New("empty commentID in url")
	ErrEmptyUsername     = errors.New("empty username in url")
	ErrRedisSetNotOk     = errors.New("redis set: result not OK")
	ErrNoAuth            = errors.New("no session found")
)
