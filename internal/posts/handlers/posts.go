package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	
	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/posts"
	"github.com/KonstantinGalanin/redditclone/internal/session"
	"github.com/KonstantinGalanin/redditclone/internal/user"
)

const (
	successMsg      = "success"
	unauthorizedMsg = "unauthorized"
	fieldPostID     = "id"
	fieldCategory   = "category"
	fieldCommentID  = "commentID"
	fieldUsername   = "username"
)

type ErrorMessage struct {
	Message string `json:"message"`
}

func WriteErrorMsg(w http.ResponseWriter, errBody string, errStatus int) {
	newError := &ErrorMessage{
		Message: errBody,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errStatus)
	if err := json.NewEncoder(w).Encode(newError); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func WriteErrorPost(w http.ResponseWriter, err error) {
	if errors.Is(err, myerrors.ErrNoPost) {
		WriteErrorMsg(w, myerrors.ErrNoPost.Error(), http.StatusNotFound)
	} else {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteResponsePost(w http.ResponseWriter, post *posts.Post, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteResponsePosts(w http.ResponseWriter, posts []*posts.Post, status int) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(posts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type PostParams struct {
	PostID    string
	Category  string
	CommentID string
	Username  string
}

func getFieldFromURL(r *http.Request, field string) (string, error) {
	vars := mux.Vars(r)
	postID, ok := vars[field]
	if !ok {
		return "", fmt.Errorf("empty %v field in url", field)
	}
	return postID, nil
}

type PostsHandler struct {
	PostsRepo      posts.PostRepo
	UserRepo       user.UserRepo
	SessionManager session.SessionManager
}

func (p *PostsHandler) getUserFromCtx(r *http.Request) (*user.User, error) {
	ctx := r.Context()
	sess, ok := ctx.Value("session").(*session.Session)
	if !ok || sess == nil {
		return nil, myerrors.ErrNoAuth
	}

	user, err := p.UserRepo.GetUserByUsername(sess.Username)
	if err != nil {
		return nil, fmt.Errorf("get user from sess: %w", err)
	}
	return user, nil
}

func (p *PostsHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	posts, err := p.PostsRepo.GetAllPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteResponsePosts(w, posts, http.StatusOK)
}

func (p *PostsHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Category string `json:"category"`
		Title    string `json:"title"`
		Type     string `json:"type"`
		URL      string `json:"url,omitempty"`
		Text     string `json:"text,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := p.getUserFromCtx(r)
	if err != nil {
		WriteErrorMsg(w, fmt.Errorf("create post %s", unauthorizedMsg).Error(), http.StatusUnauthorized)
		return
	}

	post, err := p.PostsRepo.CreatePost(data.Category, data.Title, data.Type, data.URL, data.Text, user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteResponsePost(w, post, http.StatusCreated)
}

func (p *PostsHandler) GetPost(w http.ResponseWriter, r *http.Request) {
	postID, err := getFieldFromURL(r, fieldPostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := p.PostsRepo.GetPost(postID)
	if err != nil {
		WriteErrorPost(w, err)
		return
	}

	WriteResponsePost(w, post, http.StatusOK)
}

func (p *PostsHandler) DeletePost(w http.ResponseWriter, r *http.Request) {
	postID, err := getFieldFromURL(r, fieldPostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := p.getUserFromCtx(r)
	if err != nil {
		WriteErrorMsg(w, fmt.Errorf("delete post %s", unauthorizedMsg).Error(), http.StatusUnauthorized)
		return
	}

	if err := p.PostsRepo.DeletePost(postID, user.ID); err != nil {
		WriteErrorPost(w, err)
		return
	}

	msg := &ErrorMessage{
		Message: successMsg,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (p *PostsHandler) GetByCategory(w http.ResponseWriter, r *http.Request) {
	category, err := getFieldFromURL(r, fieldCategory)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	posts, err := p.PostsRepo.GetPostsByCategory(category)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteResponsePosts(w, posts, http.StatusOK)
}

func (p *PostsHandler) CreateComment(w http.ResponseWriter, r *http.Request) {
	postID, err := getFieldFromURL(r, fieldPostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var data struct {
		Comment string `json:"comment"`
	}
	if err = json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	user, err := p.getUserFromCtx(r)
	if err != nil {
		WriteErrorMsg(w, fmt.Errorf("create comment %s", unauthorizedMsg).Error(), http.StatusUnauthorized)
		return
	}

	commentText := data.Comment
	post, err := p.PostsRepo.CreateComment(postID, commentText, user)
	if err != nil {
		WriteErrorPost(w, err)
		return
	}

	WriteResponsePost(w, post, http.StatusCreated)
}

func (p *PostsHandler) DeleteComment(w http.ResponseWriter, r *http.Request) {
	postID, err := getFieldFromURL(r, fieldPostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	commentID, err := getFieldFromURL(r, fieldCommentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	_, err = p.getUserFromCtx(r)
	if err != nil {
		WriteErrorMsg(w, fmt.Errorf("delete comment %s", unauthorizedMsg).Error(), http.StatusUnauthorized)
		return
	}

	post, err := p.PostsRepo.DeleteComment(postID, commentID)
	if err != nil {
		WriteErrorPost(w, err)
		return
	}

	WriteResponsePost(w, post, http.StatusOK)
}

func (p *PostsHandler) UpvotePost(w http.ResponseWriter, r *http.Request) {
	postID, err := getFieldFromURL(r, fieldPostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := p.getUserFromCtx(r)
	if err != nil {
		WriteErrorMsg(w, fmt.Errorf("upvote post %s", unauthorizedMsg).Error(), http.StatusUnauthorized)
		return
	}

	post, err := p.PostsRepo.UpvotePost(postID, user.ID)
	if err != nil {
		WriteErrorPost(w, err)
		return
	}

	WriteResponsePost(w, post, http.StatusOK)
}

func (p *PostsHandler) UnvotePost(w http.ResponseWriter, r *http.Request) {
	postID, err := getFieldFromURL(r, fieldPostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := p.getUserFromCtx(r)
	if err != nil {
		WriteErrorMsg(w, fmt.Errorf("unovte post %s", unauthorizedMsg).Error(), http.StatusUnauthorized)
		return
	}

	post, err := p.PostsRepo.UnvotePost(postID, user.ID)
	if err != nil {
		WriteErrorPost(w, err)
		return
	}

	WriteResponsePost(w, post, http.StatusOK)
}

func (p *PostsHandler) DownvotePost(w http.ResponseWriter, r *http.Request) {
	postID, err := getFieldFromURL(r, fieldPostID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := p.getUserFromCtx(r)
	if err != nil {
		WriteErrorMsg(w, fmt.Errorf("downvote post %s", unauthorizedMsg).Error(), http.StatusUnauthorized)
		return
	}

	post, err := p.PostsRepo.DownvotePost(postID, user.ID)
	if err != nil {
		WriteErrorPost(w, err)
		return
	}

	WriteResponsePost(w, post, http.StatusOK)
}

func (p *PostsHandler) PostsByUser(w http.ResponseWriter, r *http.Request) {
	username, err := getFieldFromURL(r, fieldUsername)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	posts, err := p.PostsRepo.GetPostsByUser(username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	WriteResponsePosts(w, posts, http.StatusOK)
}
