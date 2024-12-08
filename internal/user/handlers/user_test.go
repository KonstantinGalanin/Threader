package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/session"
	mockSession "github.com/KonstantinGalanin/redditclone/internal/session/mock"
	"github.com/KonstantinGalanin/redditclone/internal/token_manager/jwt"
	"github.com/KonstantinGalanin/redditclone/internal/token_manager/mock"
	"github.com/KonstantinGalanin/redditclone/internal/user"
	"github.com/KonstantinGalanin/redditclone/internal/user/repository"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	// mockSession "gitlab.vk-golang.ru/vk-golang/lectures/06_databases/99_hw/redditclone/internal/session/mock"
)

const (
	password = "password"
	id       = "1"
	username = "username"
)

type mockResponseWriter struct {
	HeaderMap http.Header
	Code      int
}

func (m *mockResponseWriter) Header() http.Header {
	return m.HeaderMap
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return 0, fmt.Errorf("mock write error")
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.Code = statusCode
}

func TestWriteSignupError(t *testing.T) {
	recorder := httptest.NewRecorder()
	location := "body"
	param := UsernameField
	value := "User"
	msg := myerrors.ErrUserExist.Error()

	WriteSignupError(recorder, location, param, value, msg)

	assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestWriteSignupErrorInternalError(t *testing.T) {
	response := &mockResponseWriter{
		HeaderMap: make(http.Header),
	}
	location := "body"
	param := "username"
	value := "User"
	msg := myerrors.ErrUserExist.Error()

	WriteSignupError(response, location, param, value, msg)
	assert.Equal(t, response.Code, http.StatusInternalServerError)
}

func TestWriteLoginError(t *testing.T) {
	recorder := httptest.NewRecorder()
	message := "user not found"

	WriteLoginError(recorder, message)

	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
}

func TestWriteLoginErrorInternalError(t *testing.T) {
	response := &mockResponseWriter{
		HeaderMap: make(http.Header),
	}

	WriteLoginError(response, "invalid")
	assert.Equal(t, response.Code, http.StatusInternalServerError)
}

func TestValidate(t *testing.T) {
	type user struct {
		username string
		password string
		expected error
	}
	cases := []user{
		{
			username: "User",
			password: password,
			expected: nil,
		},
		{
			username: "User",
			password: "p",
			expected: myerrors.ErrNeedMoreChars,
		},
		{
			username: "<><{};';",
			password: password,
			expected: myerrors.ErrInvalidChars,
		},
	}

	for _, c := range cases {
		err := Validate(c.username, c.password)
		assert.Equal(t, c.expected, err)
	}
}

type UserBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := repository.NewMockUserRepo(ctrl)
	sessionManager := mockSession.NewMockSessionManager(ctrl)
	jwtManager := jwt.NewJwtService()
	service := &UserHandler{
		UserRepo:       userRepo,
		SessionManager: sessionManager,
		JwtService:     jwtManager,
	}

	body, err := json.Marshal(UserBody{
		Username: username,
		Password: password,
	})
	if err != nil {
		assert.Error(t, err, "marshalling error")
	}

	cases := []struct {
		name          string
		userExpect    func()
		sessionExpect func()
		req           *http.Request
		statusCode    int
	}{
		{
			name: "no user",
			userExpect: func() {
				userRepo.EXPECT().Login(username, password).Return(&user.User{ID: id, Username: username, Password: password}, myerrors.ErrNoUser)
			},
			req:        httptest.NewRequest("POST", "/login", bytes.NewReader(body)),
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "no user",
			userExpect: func() {
				userRepo.EXPECT().Login(username, password).Return(&user.User{ID: id, Username: username, Password: password}, myerrors.ErrBadPass)
			},
			req:        httptest.NewRequest("POST", "/login", bytes.NewReader(body)),
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "no user",
			userExpect: func() {
				userRepo.EXPECT().Login(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
			},
			sessionExpect: func() {
				sessionManager.EXPECT().Create(&session.Session{
					Username: username,
				}).Return(&session.SessionID{
					ID: "1",
				}, errors.New("session error"))
			},
			req:        httptest.NewRequest("POST", "/login", bytes.NewReader(body)),
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success",
			userExpect: func() {
				userRepo.EXPECT().Login(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
			},
			sessionExpect: func() {
				sessionManager.EXPECT().Create(&session.Session{
					Username: username,
				}).Return(&session.SessionID{
					ID: "1",
				}, nil)
			},
			req:        httptest.NewRequest("POST", "/login", bytes.NewReader(body)),
			statusCode: http.StatusOK,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if c.userExpect != nil {
				c.userExpect()
			}
			if c.sessionExpect != nil {
				c.sessionExpect()
			}
			service.Login(recorder, c.req)
			assert.Equal(t, c.statusCode, recorder.Code)
		})
	}

	t.Run("decode error", func(t *testing.T) {
		response := &mockResponseWriter{
			HeaderMap: make(http.Header),
		}
		req := httptest.NewRequest("POST", "/login", nil)

		service.Login(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})
	t.Run("validate error", func(t *testing.T) {
		username := "/.<>:;"
		password := "p"

		bodyValid, err := json.Marshal(UserBody{
			Username: username,
			Password: password,
		})
		if err != nil {
			assert.Error(t, err, "marshalling error")
		}

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(bodyValid))

		service.Login(recorder, req)

		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
	})

	t.Run("jwt error", func(t *testing.T) {
		userRepoJWT := repository.NewMockUserRepo(ctrl)
		sessionManagerJWT := mockSession.NewMockSessionManager(ctrl)
		jwtManagerJWT := mock.NewMockTokenManager(ctrl)
		serviceJWT := &UserHandler{
			UserRepo:       userRepoJWT,
			SessionManager: sessionManagerJWT,
			JwtService:     jwtManagerJWT,
		}
		userRepoJWT.EXPECT().Login(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
		jwtManagerJWT.EXPECT().CreateToken(&user.User{ID: id, Username: username, Password: password}).Return([]byte(""), errors.New("some error"))

		response := &mockResponseWriter{
			HeaderMap: make(http.Header),
		}
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))

		serviceJWT.Login(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("login write error", func(t *testing.T) {
		response := &mockResponseWriter{
			HeaderMap: make(http.Header),
		}
		userRepo.EXPECT().Login(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
		sessionManager.EXPECT().Create(&session.Session{
			Username: username,
		}).Return(&session.SessionID{
			ID: "1",
		}, nil)
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))

		service.Login(response, req)

		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestSignup(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userRepo := repository.NewMockUserRepo(ctrl)
	sessionManager := mockSession.NewMockSessionManager(ctrl)
	jwtManager := jwt.NewJwtService()
	service := &UserHandler{
		UserRepo:       userRepo,
		SessionManager: sessionManager,
		JwtService:     jwtManager,
	}

	body, err := json.Marshal(UserBody{
		Username: username,
		Password: password,
	})
	if err != nil {
		assert.Error(t, err, "marshalling error")
	}

	cases := []struct {
		name          string
		userExpect    func()
		sessionExpect func()
		req           *http.Request
		statusCode    int
	}{
		{
			name: "no user",
			userExpect: func() {
				userRepo.EXPECT().Signup(username, password).Return(&user.User{ID: id, Username: username, Password: password}, myerrors.ErrUserExist)
			},
			req:        httptest.NewRequest("POST", "/signup", bytes.NewReader(body)),
			statusCode: http.StatusUnprocessableEntity,
		},
		{
			name: "session error",
			userExpect: func() {
				userRepo.EXPECT().Signup(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
			},
			sessionExpect: func() {
				sessionManager.EXPECT().Create(&session.Session{
					Username: username,
				}).Return(&session.SessionID{
					ID: "1",
				}, errors.New("session error"))
			},
			req:        httptest.NewRequest("POST", "/signup", bytes.NewReader(body)),
			statusCode: http.StatusInternalServerError,
		},
		{
			name: "success",
			userExpect: func() {
				userRepo.EXPECT().Signup(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
			},
			sessionExpect: func() {
				sessionManager.EXPECT().Create(&session.Session{
					Username: username,
				}).Return(&session.SessionID{
					ID: "1",
				}, nil)
			},
			req:        httptest.NewRequest("POST", "/signup", bytes.NewReader(body)),
			statusCode: http.StatusCreated,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if c.userExpect != nil {
				c.userExpect()
			}
			if c.sessionExpect != nil {
				c.sessionExpect()
			}
			service.Signup(recorder, c.req)
			assert.Equal(t, c.statusCode, recorder.Code)
		})
	}

	t.Run("validate error", func(t *testing.T) {
		username := "/.<>:;"
		password := "p"
		bodyValid, err := json.Marshal(UserBody{
			Username: username,
			Password: password,
		})
		if err != nil {
			assert.Error(t, err, "marshalling error")
		}
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/signup", bytes.NewReader(bodyValid))
		service.Signup(recorder, req)

		assert.Equal(t, http.StatusUnprocessableEntity, recorder.Code)
	})

	t.Run("write error", func(t *testing.T) {
		userRepo.EXPECT().Signup(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
		sessionManager.EXPECT().Create(&session.Session{
			Username: username,
		}).Return(&session.SessionID{
			ID: "1",
		}, nil)

		response := &mockResponseWriter{
			HeaderMap: make(http.Header),
		}
		req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))

		service.Signup(response, req)

		assert.Equal(t, http.StatusInternalServerError, response.Code)

	})

	t.Run("decode error", func(t *testing.T) {
		response := &mockResponseWriter{
			HeaderMap: make(http.Header),
		}
		req := httptest.NewRequest("POST", "/login", nil)

		service.Signup(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("jwt error", func(t *testing.T) {
		userRepo := repository.NewMockUserRepo(ctrl)
		sessionManager := mockSession.NewMockSessionManager(ctrl)
		jwtManager := mock.NewMockTokenManager(ctrl)
		serviceJWT := &UserHandler{
			UserRepo:       userRepo,
			SessionManager: sessionManager,
			JwtService:     jwtManager,
		}
		userRepo.EXPECT().Signup(username, password).Return(&user.User{ID: id, Username: username, Password: password}, nil)
		jwtManager.EXPECT().CreateToken(&user.User{ID: id, Username: username, Password: password}).Return([]byte(""), errors.New("some error"))

		response := &mockResponseWriter{
			HeaderMap: make(http.Header),
		}
		req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))

		serviceJWT.Signup(response, req)

		assert.Equal(t, http.StatusBadRequest, response.Code)
	})
}
