package tokenmanager

import "github.com/KonstantinGalanin/redditclone/internal/user"

//go:generate mockgen -source=token_manager.go -destination=mock/jwt_mock.go -package=mock JwtService
type TokenManager interface {
	CreateToken(userItem *user.User) ([]byte, error)
}
