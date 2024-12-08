package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/session"
	"github.com/KonstantinGalanin/redditclone/internal/user"
)

var (
	usernameValid = regexp.MustCompile(`[a-zA-Z0-9]+`)
	passwordValid = regexp.MustCompile(`.{8,}`)
)

const (
	UsernameField = "username"
)

type ErrorSignup struct {
	Error []ErrorInfo `json:"errors"`
}

type ErrorInfo struct {
	Location string `json:"location"`
	Param    string `json:"param"`
	Value    string `json:"value"`
	Msg      string `json:"msg"`
}

type ErrorLogin struct {
	Message string `json:"message"`
}

func Send(w http.ResponseWriter, resp []byte, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteSignupError(w http.ResponseWriter, location, param, value, msg string) {
	newError := &ErrorSignup{
		[]ErrorInfo{
			{
				Location: location,
				Param:    param,
				Value:    value,
				Msg:      msg,
			},
		},
	}
	resp, err := json.Marshal(newError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	Send(w, resp, http.StatusUnprocessableEntity)
}

func WriteLoginError(w http.ResponseWriter, message string) {
	newError := &ErrorLogin{
		Message: message,
	}
	resp, err := json.Marshal(newError)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	Send(w, resp, http.StatusUnauthorized)
}

func Validate(username, password string) error {
	if !usernameValid.MatchString(username) {
		return myerrors.ErrInvalidChars
	}

	if !passwordValid.MatchString(password) {
		return myerrors.ErrNeedMoreChars
	}

	return nil
}

type JwtService interface {
	CreateToken(userItem *user.User) ([]byte, error)
}

type UserHandler struct {
	UserRepo       user.UserRepo
	SessionManager session.SessionManager
	JwtService     JwtService
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := Validate(data.Username, data.Password); err != nil {
		WriteLoginError(w, err.Error())
		return
	}

	userItem, err := h.UserRepo.Login(data.Username, data.Password)
	if errors.Is(err, myerrors.ErrNoUser) {
		WriteLoginError(w, myerrors.ErrNoUser.Error())
		return
	}
	if errors.Is(err, myerrors.ErrBadPass) {
		WriteLoginError(w, myerrors.ErrBadPass.Error())
		return
	}

	resp, err := h.JwtService.CreateToken(userItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := h.SessionManager.Create(&session.Session{
		Username: data.Username,
	})
	if err != nil {
		http.Error(w, fmt.Errorf("cant create session: %w", err).Error(), http.StatusInternalServerError)
	}
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   sess.ID,
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, &cookie)

	Send(w, resp, http.StatusOK)
}

func (h *UserHandler) Signup(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Password string `json:"password"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := Validate(data.Username, data.Password); err != nil {
		WriteSignupError(w, "body", UsernameField, data.Username, err.Error())
		return
	}

	userItem, err := h.UserRepo.Signup(data.Username, data.Password)
	if errors.Is(err, myerrors.ErrUserExist) {
		WriteSignupError(w, "body", UsernameField, data.Username, myerrors.ErrUserExist.Error())
		return
	}

	resp, err := h.JwtService.CreateToken(userItem)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	sess, err := h.SessionManager.Create(&session.Session{
		Username: data.Username,
	})
	if err != nil {
		http.Error(w, fmt.Errorf("cant create session: %w", err).Error(), http.StatusInternalServerError)
	}
	cookie := http.Cookie{
		Name:    "session_id",
		Value:   sess.ID,
		Expires: time.Now().Add(24 * time.Hour),
	}
	http.SetCookie(w, &cookie)

	Send(w, resp, http.StatusCreated)
}
