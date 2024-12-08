package posts

import (
	"time"

	"github.com/KonstantinGalanin/redditclone/internal/user"
)

type Post struct {
	Author           *user.User `json:"author" bson:"author"`
	Category         string     `json:"category" bson:"category"`
	Comments         []*Comment `json:"comments" bson:"comments"`
	Created          time.Time  `json:"created" bson:"created"`
	ID               string     `json:"id" bson:"_id"`
	Score            int        `json:"score" bson:"score"`
	Title            string     `json:"title" bson:"title"`
	Type             string     `json:"type" bson:"type"`
	UpvotePercentage int        `json:"upvotePercentage" bson:"upvotePercentage"`
	URL              string     `json:"url,omitempty" bson:"url,omitempty"`
	Text             string     `json:"text,omitempty" bson:"text,omitempty"`
	Views            int        `json:"views" bson:"views"`
	Votes            []*Vote    `json:"votes" bson:"votes"`
}

type Comment struct {
	Author  *user.User `json:"author" bson:"author"`
	Body    string     `json:"body" bson:"body"`
	Created time.Time  `json:"created" bson:"created"`
	ID      string     `json:"id" bson:"_id"`
}

type Vote struct {
	UserID string `json:"user" bson:"user"`
	Vote   int    `json:"vote" bson:"vote"`
}

//go:generate mockgen -source=posts.go -destination=repository/repo_mock.go -package=repository PostRepo
type PostRepo interface {
	GetAllPosts() ([]*Post, error)
	CreatePost(category, title, typePost, url, text string, author *user.User) (*Post, error)
	GetPost(postID string) (*Post, error)
	GetPostsByCategory(category string) ([]*Post, error)
	CreateComment(postID, text string, author *user.User) (*Post, error)
	DeleteComment(postID string, commentID string) (*Post, error)
	UpvotePost(postID, userID string) (*Post, error)
	UnvotePost(postID, userID string) (*Post, error)
	DownvotePost(postID, userID string) (*Post, error)
	DeletePost(postID, userID string) error
	GetPostsByUser(username string) ([]*Post, error)
}
