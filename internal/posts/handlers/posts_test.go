package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KonstantinGalanin/redditclone/internal/session"
	"github.com/KonstantinGalanin/redditclone/internal/session/mock"
	"github.com/KonstantinGalanin/redditclone/internal/user"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/posts"

	repositoryPosts "github.com/KonstantinGalanin/redditclone/internal/posts/repository"
	repositoryUser "github.com/KonstantinGalanin/redditclone/internal/user/repository"
)

const (
	username  = "User"
	id        = "1"
	password  = "password"
	postID    = "1"
	category  = "music"
	commentID = "2"
)

var expectedUser = &user.User{
	Username: username,
	Password: password,
	ID:       id,
}

type mockResponseWriter struct {
	HeaderMap http.Header
	Code      int
}

func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		HeaderMap: make(http.Header),
	}
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

func TestWriteErrorMsg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		WriteErrorMsg(recorder, "some error", http.StatusBadRequest)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	})

	t.Run("error", func(t *testing.T) {
		response := newMockResponseWriter()
		WriteErrorMsg(response, "some error", http.StatusBadRequest)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestWriteErrorPost(t *testing.T) {
	t.Run("no post", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		WriteErrorPost(recorder, myerrors.ErrNoPost)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
		assert.Equal(t, http.StatusNotFound, recorder.Code)
	})
	t.Run("bad pass", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		WriteErrorPost(recorder, myerrors.ErrBadPass)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})
}

func TestWriteResponsePost(t *testing.T) {
	post := &posts.Post{}

	t.Run("success", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		WriteResponsePost(recorder, post, http.StatusBadRequest)
		assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	})

	t.Run("error", func(t *testing.T) {
		response := newMockResponseWriter()
		WriteResponsePost(response, post, http.StatusBadRequest)
		assert.Equal(t, http.StatusInternalServerError, response.Code)
	})
}

func TestWriteResponsePosts(t *testing.T) {
	posts := []*posts.Post{}
	// success
	recorder := httptest.NewRecorder()
	WriteResponsePosts(recorder, posts, http.StatusBadRequest)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

	// error
	response := newMockResponseWriter()
	WriteResponsePosts(response, posts, http.StatusBadRequest)
	assert.Equal(t, http.StatusInternalServerError, response.Code)
}

func TestGetFieldFromURL(t *testing.T) {
	router := mux.NewRouter()

	router.HandleFunc("/test/{test_var}", func(w http.ResponseWriter, r *http.Request) {
		nameVar := "test_var"
		expected := "1234"

		value, err := getFieldFromURL(r, nameVar)

		assert.NoError(t, err)
		assert.Equal(t, expected, value)
	})

	req := httptest.NewRequest("GET", "/test/1234", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
}

func TestGetFieldFromURLNoField(t *testing.T) {
	router := mux.NewRouter()

	router.HandleFunc("/test/{test_var}", func(w http.ResponseWriter, r *http.Request) {
		nameVar := "invalid_field"

		value, err := getFieldFromURL(r, nameVar)

		assert.Equal(t, fmt.Errorf("empty %v field in url", nameVar), err)
		assert.Equal(t, "", value)
	})

	req := httptest.NewRequest("GET", "/test/1234", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)
}

func newMockService(postRepo posts.PostRepo, userRepo user.UserRepo, sessionManager session.SessionManager) *PostsHandler {
	return &PostsHandler{
		PostsRepo:      postRepo,
		UserRepo:       userRepo,
		SessionManager: sessionManager,
	}
}

func TestGetUserFromCtxNoSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.WithValue(context.Background(), "session", "invalid value")
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	user, err := service.getUserFromCtx(req)

	assert.Error(t, err)
	assert.Equal(t, err, fmt.Errorf("no session found"))
	assert.Nil(t, user)
}

func TestGetUserFromCtxNoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.WithValue(context.Background(), "session", &session.Session{
		Username: username,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)
	userRepo.EXPECT().GetUserByUsername(username).Return(nil, errors.New("no user session error"))

	user, err := service.getUserFromCtx(req)

	assert.Error(t, err)
	assert.Equal(t, err.Error(), fmt.Errorf("get user from sess: no user session error").Error())
	assert.Nil(t, user)
}

func TestGetUserFromCtxSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.WithValue(context.Background(), "session", &session.Session{
		Username: username,
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)
	userRepo.EXPECT().GetUserByUsername(username).Return(expectedUser, nil)

	actualUser, err := service.getUserFromCtx(req)

	assert.NoError(t, err)
	assert.Equal(t, actualUser, expectedUser)
}

func TestGetAllError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	cases := []struct {
		name       string
		statusCode int
		postExpect func()
		userExpect func()
	}{
		{
			name:       "get all posts error",
			statusCode: http.StatusInternalServerError,
			postExpect: func() {
				postsRepo.EXPECT().GetAllPosts().Return(nil, errors.New("some error"))
			},
		},
		{
			name:       "success",
			statusCode: http.StatusOK,
			postExpect: func() {
				postsRepo.EXPECT().GetAllPosts().Return([]*posts.Post{}, nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.GetAll(recorder, req)
			assert.Equal(t, c.statusCode, recorder.Code)
		})
	}
}

type DataBody struct {
	Category string `json:"category"`
	Title    string `json:"title"`
	Type     string `json:"type"`
	URL      string `json:"url,omitempty"`
	Text     string `json:"text,omitempty"`
}

func TestCreatePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	data := DataBody{
		Category: "category",
		Title:    "title",
		Type:     "type",
		Text:     "Text",
	}
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshalling error %v", err)
	}

	cases := []struct {
		name       string
		statusCode int
		postExpect func()
		userExpect func()
		req        *http.Request
	}{
		{
			name:       "wrong responder",
			statusCode: http.StatusInternalServerError,
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:       "error get user from context",
			statusCode: http.StatusUnauthorized,
			req:        httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)),
		},
		{
			name:       "create post internal error",
			statusCode: http.StatusInternalServerError,
			req: httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})),
			postExpect: func() {
				postsRepo.EXPECT().CreatePost(data.Category, data.Title, data.Type, data.URL, data.Text, expectedUser).Return(nil, errors.New("some error"))
			},
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
		},
		{
			name:       "success",
			statusCode: http.StatusCreated,
			req: httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})),
			postExpect: func() {
				postsRepo.EXPECT().CreatePost(data.Category, data.Title, data.Type, data.URL, data.Text, expectedUser).Return(&posts.Post{}, nil)
			},
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.CreatePost(recorder, c.req)
			assert.Equal(t, c.statusCode, recorder.Code)
		})
	}
}

func TestGetPost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	cases := []struct {
		name       string
		statusCode int
		postExpect func()
		userExpect func()
		req        *http.Request
	}{
		{
			name:       "wrong responder",
			statusCode: http.StatusBadRequest,
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:       "get field from url error",
			statusCode: http.StatusBadRequest,
			req:        httptest.NewRequest(http.MethodGet, "/", nil),
		},
		{
			name:       "get post internal error",
			statusCode: http.StatusInternalServerError,
			req:        mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID}),
			postExpect: func() {
				postsRepo.EXPECT().GetPost(postID).Return(nil, errors.New("some error"))
			},
		},
		{
			name:       "success",
			statusCode: http.StatusOK,
			req:        mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID}),
			postExpect: func() {
				postsRepo.EXPECT().GetPost(postID).Return(&posts.Post{}, nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.GetPost(recorder, c.req)
			assert.Equal(t, c.statusCode, recorder.Code)
		})
	}
}

func TestDeletePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	data := DataBody{
		Category: "category",
		Title:    "title",
		Type:     "type",
		Text:     "Text",
	}
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshalling error %v", err)
	}

	cases := []struct {
		name         string
		statusCode   int
		postExpect   func()
		userExpect   func()
		req          *http.Request
		mockRecorder bool
	}{
		{
			name:         "get field from url error",
			statusCode:   http.StatusBadRequest,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			mockRecorder: false,
		},
		{
			name:         "error get user from context",
			statusCode:   http.StatusUnauthorized,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)), map[string]string{"id": postID}),
			mockRecorder: false,
		},
		{
			name:       "delete post internal error",
			statusCode: http.StatusInternalServerError,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID}),
			postExpect: func() {
				postsRepo.EXPECT().DeletePost(postID, expectedUser.ID).Return(errors.New("some error"))
			},
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
		},
		{
			name:       "delete post write error",
			statusCode: http.StatusInternalServerError,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID}),
			postExpect: func() {
				postsRepo.EXPECT().DeletePost(postID, expectedUser.ID).Return(nil)
			},
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
			mockRecorder: true,
		},
		{
			name:       "success",
			statusCode: http.StatusOK,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID}),
			postExpect: func() {
				postsRepo.EXPECT().DeletePost(postID, expectedUser.ID).Return(nil)
			},
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var recorder http.ResponseWriter
			recorder = httptest.NewRecorder()
			if c.mockRecorder {
				recorder = newMockResponseWriter()
			}

			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.DeletePost(recorder, c.req)

			if c.mockRecorder {
				res := recorder.(*mockResponseWriter)
				assert.Equal(t, c.statusCode, res.Code)
			} else {
				res := recorder.(*httptest.ResponseRecorder)
				assert.Equal(t, c.statusCode, res.Code)
			}
		})
	}
}

func TestGetByCategory(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	cases := []struct {
		name         string
		statusCode   int
		postExpect   func()
		userExpect   func()
		req          *http.Request
		mockRecorder bool
	}{
		{
			name:         "get field from url error",
			statusCode:   http.StatusBadRequest,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			mockRecorder: false,
		},
		{
			name:       "delete post internal error",
			statusCode: http.StatusInternalServerError,
			req:        mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"category": category}),
			postExpect: func() {
				postsRepo.EXPECT().GetPostsByCategory(category).Return([]*posts.Post{}, errors.New("some error"))
			},
			mockRecorder: false,
		},
		{
			name:       "success",
			statusCode: http.StatusOK,
			req:        mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"category": category}),
			postExpect: func() {
				postsRepo.EXPECT().GetPostsByCategory(category).Return([]*posts.Post{}, nil)
			},
			mockRecorder: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var recorder http.ResponseWriter
			recorder = httptest.NewRecorder()
			if c.mockRecorder {
				recorder = newMockResponseWriter()
			}

			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.GetByCategory(recorder, c.req)

			if c.mockRecorder {
				res := recorder.(*mockResponseWriter)
				assert.Equal(t, c.statusCode, res.Code)
			} else {
				res := recorder.(*httptest.ResponseRecorder)
				assert.Equal(t, c.statusCode, res.Code)
			}
		})
	}
}

type DataComment struct {
	Comment string `json:"comment"`
}

func TestCreateComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	data := DataComment{
		Comment: "comment",
	}
	body, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshalling error %vv", err)
	}

	cases := []struct {
		name         string
		statusCode   int
		postExpect   func()
		userExpect   func()
		req          *http.Request
		mockRecorder bool
	}{
		{
			name:         "get field from url error",
			statusCode:   http.StatusBadRequest,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			mockRecorder: false,
		},
		{
			name:         "wrong responder",
			statusCode:   http.StatusInternalServerError,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID}),
			mockRecorder: true,
		},
		{
			name:         "error get user from context",
			statusCode:   http.StatusUnauthorized,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)), map[string]string{"id": postID}),
			mockRecorder: false,
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID}),
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
			postExpect: func() {
				postsRepo.EXPECT().CreateComment(postID, data.Comment, expectedUser).Return(nil, errors.New("some error"))
			},
			mockRecorder: false,
		},
		{
			name:       "success",
			statusCode: http.StatusCreated,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", bytes.NewReader(body)).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID}),
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
			postExpect: func() {
				postsRepo.EXPECT().CreateComment(postID, data.Comment, expectedUser).Return(&posts.Post{}, nil)
			},
			mockRecorder: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var recorder http.ResponseWriter
			recorder = httptest.NewRecorder()
			if c.mockRecorder {
				recorder = newMockResponseWriter()
			}

			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.CreateComment(recorder, c.req)

			if c.mockRecorder {
				res := recorder.(*mockResponseWriter)
				assert.Equal(t, c.statusCode, res.Code)
			} else {
				res := recorder.(*httptest.ResponseRecorder)
				assert.Equal(t, c.statusCode, res.Code)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	cases := []struct {
		name         string
		statusCode   int
		postExpect   func()
		userExpect   func()
		req          *http.Request
		mockRecorder bool
	}{
		{
			name:         "get post field from url error",
			statusCode:   http.StatusBadRequest,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			mockRecorder: false,
		},
		{
			name:         "get comment id field from url error",
			statusCode:   http.StatusBadRequest,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID}),
			mockRecorder: false,
		},
		{
			name:         "context error",
			statusCode:   http.StatusUnauthorized,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID, "commentID": commentID}),
			mockRecorder: false,
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID, "commentID": commentID}),
			mockRecorder: false,
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
			postExpect: func() {
				postsRepo.EXPECT().DeleteComment(postID, commentID).Return(nil, errors.New("some error"))
			},
		},
		{
			name:       "success",
			statusCode: http.StatusOK,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID, "commentID": commentID}),
			mockRecorder: false,
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
			postExpect: func() {
				postsRepo.EXPECT().DeleteComment(postID, commentID).Return(&posts.Post{}, nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var recorder http.ResponseWriter
			recorder = httptest.NewRecorder()
			if c.mockRecorder {
				recorder = newMockResponseWriter()
			}

			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.DeleteComment(recorder, c.req)

			if c.mockRecorder {
				res := recorder.(*mockResponseWriter)
				assert.Equal(t, c.statusCode, res.Code)
			} else {
				res := recorder.(*httptest.ResponseRecorder)
				assert.Equal(t, c.statusCode, res.Code)
			}
		})
	}
}

func TestUpvoteUnvotePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	cases := []struct {
		name         string
		statusCode   int
		postExpect   func()
		userExpect   func()
		req          *http.Request
		mockRecorder bool
	}{
		{
			name:         "get post field from url error",
			statusCode:   http.StatusBadRequest,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			mockRecorder: false,
		},
		{
			name:         "context error",
			statusCode:   http.StatusUnauthorized,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID}),
			mockRecorder: false,
		},
		{
			name:         "context error",
			statusCode:   http.StatusUnauthorized,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID}),
			mockRecorder: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var recorder http.ResponseWriter
			recorder = httptest.NewRecorder()
			if c.mockRecorder {
				recorder = newMockResponseWriter()
			}

			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.UpvotePost(recorder, c.req)
			service.UnvotePost(recorder, c.req)

			if c.mockRecorder {
				res := recorder.(*mockResponseWriter)
				assert.Equal(t, c.statusCode, res.Code)
			} else {
				res := recorder.(*httptest.ResponseRecorder)
				assert.Equal(t, c.statusCode, res.Code)
			}
		})
	}

	t.Run("internal error", func(t *testing.T) {
		postID := "1"
		recorder := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "session", &session.Session{
			Username: expectedUser.Username,
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"id": postID})
		userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
		postsRepo.EXPECT().UpvotePost(postID, expectedUser.ID).Return(nil, errors.New("some error"))
		service.UpvotePost(recorder, req)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("internal error", func(t *testing.T) {
		postID := "1"
		recorder := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "session", &session.Session{
			Username: expectedUser.Username,
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"id": postID})
		userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
		postsRepo.EXPECT().UnvotePost(postID, expectedUser.ID).Return(nil, errors.New("some error"))
		service.UnvotePost(recorder, req)
		assert.Equal(t, http.StatusInternalServerError, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		postID := "1"
		recorder := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "session", &session.Session{
			Username: expectedUser.Username,
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"id": postID})
		userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
		postsRepo.EXPECT().UpvotePost(postID, expectedUser.ID).Return(&posts.Post{}, nil)
		service.UpvotePost(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("success", func(t *testing.T) {
		postID := "1"
		recorder := httptest.NewRecorder()
		ctx := context.WithValue(context.Background(), "session", &session.Session{
			Username: expectedUser.Username,
		})
		req := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(ctx)
		req = mux.SetURLVars(req, map[string]string{"id": postID})
		userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
		postsRepo.EXPECT().UnvotePost(postID, expectedUser.ID).Return(&posts.Post{}, nil)
		service.UnvotePost(recorder, req)
		assert.Equal(t, http.StatusOK, recorder.Code)
	})
}

func TestDownvotePost(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	cases := []struct {
		name         string
		statusCode   int
		postExpect   func()
		userExpect   func()
		req          *http.Request
		mockRecorder bool
	}{
		{
			name:         "get post field from url error",
			statusCode:   http.StatusBadRequest,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			mockRecorder: false,
		},
		{
			name:         "context error",
			statusCode:   http.StatusUnauthorized,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"id": postID}),
			mockRecorder: false,
		},
		{
			name:       "internal error",
			statusCode: http.StatusInternalServerError,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID}),
			mockRecorder: false,
			postExpect: func() {
				postsRepo.EXPECT().DownvotePost(postID, expectedUser.ID).Return(nil, errors.New("some error"))
			},
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
		},
		{
			name:       "success",
			statusCode: http.StatusOK,
			req: mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil).WithContext(context.WithValue(context.Background(), "session", &session.Session{
				Username: expectedUser.Username,
			})), map[string]string{"id": postID}),
			mockRecorder: false,
			postExpect: func() {
				postsRepo.EXPECT().DownvotePost(postID, expectedUser.ID).Return(&posts.Post{}, nil)
			},
			userExpect: func() {
				userRepo.EXPECT().GetUserByUsername(expectedUser.Username).Return(expectedUser, nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var recorder http.ResponseWriter
			recorder = httptest.NewRecorder()
			if c.mockRecorder {
				recorder = newMockResponseWriter()
			}

			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.DownvotePost(recorder, c.req)

			if c.mockRecorder {
				res := recorder.(*mockResponseWriter)
				assert.Equal(t, c.statusCode, res.Code)
			} else {
				res := recorder.(*httptest.ResponseRecorder)
				assert.Equal(t, c.statusCode, res.Code)
			}
		})
	}
}

func TestPostsByUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	postsRepo := repositoryPosts.NewMockPostRepo(ctrl)
	userRepo := repositoryUser.NewMockUserRepo(ctrl)
	sessionManager := mock.NewMockSessionManager(ctrl)

	service := newMockService(postsRepo, userRepo, sessionManager)

	cases := []struct {
		name         string
		statusCode   int
		postExpect   func()
		userExpect   func()
		req          *http.Request
		mockRecorder bool
	}{
		{
			name:         "get post field from url error",
			statusCode:   http.StatusBadRequest,
			req:          httptest.NewRequest(http.MethodGet, "/", nil),
			mockRecorder: false,
		},
		{
			name:         "internal error",
			statusCode:   http.StatusInternalServerError,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"username": username}),
			mockRecorder: false,
			postExpect: func() {
				postsRepo.EXPECT().GetPostsByUser(expectedUser.Username).Return(nil, errors.New("some error"))

			},
		},
		{
			name:         "success",
			statusCode:   http.StatusOK,
			req:          mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{"username": username}),
			mockRecorder: false,
			postExpect: func() {
				postsRepo.EXPECT().GetPostsByUser(expectedUser.Username).Return([]*posts.Post{}, nil)
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var recorder http.ResponseWriter
			recorder = httptest.NewRecorder()
			if c.mockRecorder {
				recorder = newMockResponseWriter()
			}

			if c.postExpect != nil {
				c.postExpect()
			}
			if c.userExpect != nil {
				c.userExpect()
			}
			service.PostsByUser(recorder, c.req)

			if c.mockRecorder {
				res := recorder.(*mockResponseWriter)
				assert.Equal(t, c.statusCode, res.Code)
			} else {
				res := recorder.(*httptest.ResponseRecorder)
				assert.Equal(t, c.statusCode, res.Code)
			}
		})
	}
}
