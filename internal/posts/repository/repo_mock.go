// Code generated by MockGen. DO NOT EDIT.
// Source: posts.go

// Package repository is a generated GoMock package.
package repository

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	posts "github.com/KonstantinGalanin/redditclone/internal/posts"
	user "github.com/KonstantinGalanin/redditclone/internal/user"
)

// MockPostRepo is a mock of PostRepo interface.
type MockPostRepo struct {
	ctrl     *gomock.Controller
	recorder *MockPostRepoMockRecorder
}

// MockPostRepoMockRecorder is the mock recorder for MockPostRepo.
type MockPostRepoMockRecorder struct {
	mock *MockPostRepo
}

// NewMockPostRepo creates a new mock instance.
func NewMockPostRepo(ctrl *gomock.Controller) *MockPostRepo {
	mock := &MockPostRepo{ctrl: ctrl}
	mock.recorder = &MockPostRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPostRepo) EXPECT() *MockPostRepoMockRecorder {
	return m.recorder
}

// CreateComment mocks base method.
func (m *MockPostRepo) CreateComment(postID, text string, author *user.User) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateComment", postID, text, author)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateComment indicates an expected call of CreateComment.
func (mr *MockPostRepoMockRecorder) CreateComment(postID, text, author interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateComment", reflect.TypeOf((*MockPostRepo)(nil).CreateComment), postID, text, author)
}

// CreatePost mocks base method.
func (m *MockPostRepo) CreatePost(category, title, typePost, url, text string, author *user.User) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreatePost", category, title, typePost, url, text, author)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreatePost indicates an expected call of CreatePost.
func (mr *MockPostRepoMockRecorder) CreatePost(category, title, typePost, url, text, author interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreatePost", reflect.TypeOf((*MockPostRepo)(nil).CreatePost), category, title, typePost, url, text, author)
}

// DeleteComment mocks base method.
func (m *MockPostRepo) DeleteComment(postID, commentID string) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteComment", postID, commentID)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteComment indicates an expected call of DeleteComment.
func (mr *MockPostRepoMockRecorder) DeleteComment(postID, commentID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteComment", reflect.TypeOf((*MockPostRepo)(nil).DeleteComment), postID, commentID)
}

// DeletePost mocks base method.
func (m *MockPostRepo) DeletePost(postID, userID string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePost", postID, userID)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePost indicates an expected call of DeletePost.
func (mr *MockPostRepoMockRecorder) DeletePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePost", reflect.TypeOf((*MockPostRepo)(nil).DeletePost), postID, userID)
}

// DownvotePost mocks base method.
func (m *MockPostRepo) DownvotePost(postID, userID string) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DownvotePost", postID, userID)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DownvotePost indicates an expected call of DownvotePost.
func (mr *MockPostRepoMockRecorder) DownvotePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DownvotePost", reflect.TypeOf((*MockPostRepo)(nil).DownvotePost), postID, userID)
}

// GetAllPosts mocks base method.
func (m *MockPostRepo) GetAllPosts() ([]*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllPosts")
	ret0, _ := ret[0].([]*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllPosts indicates an expected call of GetAllPosts.
func (mr *MockPostRepoMockRecorder) GetAllPosts() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllPosts", reflect.TypeOf((*MockPostRepo)(nil).GetAllPosts))
}

// GetPost mocks base method.
func (m *MockPostRepo) GetPost(postID string) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPost", postID)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPost indicates an expected call of GetPost.
func (mr *MockPostRepoMockRecorder) GetPost(postID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPost", reflect.TypeOf((*MockPostRepo)(nil).GetPost), postID)
}

// GetPostsByCategory mocks base method.
func (m *MockPostRepo) GetPostsByCategory(category string) ([]*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPostsByCategory", category)
	ret0, _ := ret[0].([]*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPostsByCategory indicates an expected call of GetPostsByCategory.
func (mr *MockPostRepoMockRecorder) GetPostsByCategory(category interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPostsByCategory", reflect.TypeOf((*MockPostRepo)(nil).GetPostsByCategory), category)
}

// GetPostsByUser mocks base method.
func (m *MockPostRepo) GetPostsByUser(username string) ([]*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPostsByUser", username)
	ret0, _ := ret[0].([]*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPostsByUser indicates an expected call of GetPostsByUser.
func (mr *MockPostRepoMockRecorder) GetPostsByUser(username interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPostsByUser", reflect.TypeOf((*MockPostRepo)(nil).GetPostsByUser), username)
}

// UnvotePost mocks base method.
func (m *MockPostRepo) UnvotePost(postID, userID string) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UnvotePost", postID, userID)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UnvotePost indicates an expected call of UnvotePost.
func (mr *MockPostRepoMockRecorder) UnvotePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UnvotePost", reflect.TypeOf((*MockPostRepo)(nil).UnvotePost), postID, userID)
}

// UpvotePost mocks base method.
func (m *MockPostRepo) UpvotePost(postID, userID string) (*posts.Post, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpvotePost", postID, userID)
	ret0, _ := ret[0].(*posts.Post)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpvotePost indicates an expected call of UpvotePost.
func (mr *MockPostRepoMockRecorder) UpvotePost(postID, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpvotePost", reflect.TypeOf((*MockPostRepo)(nil).UpvotePost), postID, userID)
}
