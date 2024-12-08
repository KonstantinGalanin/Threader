package jwt

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/KonstantinGalanin/redditclone/internal/user"
	jwtToken "github.com/dgrijalva/jwt-go"
)

const (
	ExpTime = 7 * 24 * 60 * 60
)

type JwtInfo struct {
	User *user.User `json:"user"`
	Iat  int64      `json:"iat"`
	Exp  int64      `json:"exp"`
	jwtToken.StandardClaims
}

var (
	TokenSecret = []byte(os.Getenv("TOKEN_SECRET"))
)

type JwtService struct{}

func NewJwtService() *JwtService {
	return &JwtService{}
}

func (j *JwtService) CreateToken(userItem *user.User) ([]byte, error) {
	token := jwtToken.NewWithClaims(jwtToken.SigningMethodHS256, jwtToken.MapClaims{
		"user": userItem,
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Unix() + ExpTime,
	})

	tokenString, err := token.SignedString(TokenSecret)
	if err != nil {
		return nil, err
	}

	resp, err := json.Marshal(map[string]interface{}{
		"token": tokenString,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetToken(tokenString string) (*user.User, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	claims := &JwtInfo{}
	token, err := jwtToken.ParseWithClaims(tokenString, claims, func(t *jwtToken.Token) (interface{}, error) {
		return TokenSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("Unauthorized")
	}

	return claims.User, nil
}
