package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/KonstantinGalanin/redditclone/internal/myerrors"
	"github.com/KonstantinGalanin/redditclone/internal/posts"
	"github.com/KonstantinGalanin/redditclone/internal/user"
)


const (
	LIKE                  = 1
	DISLIKE               = -1
	InitialScore          = 1
	ZeroViews             = 0
	ZeroPercent           = 0
	FullPercent           = 100
	TimeoutVal            = 10
	NoCredentialsToDelete = "No cretdentials to delete post: "
)

type PostMongoDB struct {
	db *mongo.Collection
}

func NewPostMongoDB(db *mongo.Collection) *PostMongoDB {
	return &PostMongoDB{
		db: db,
	}
}

func (p *PostMongoDB) withTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), TimeoutVal*time.Second)
}

func (p *PostMongoDB) GetAllPosts() ([]*posts.Post, error) {
	posts := []*posts.Post{}
	ctx, cancel := p.withTimeout()
	defer cancel()
	c, err := p.db.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("mongodb get all posts: %w", err)
	}
	err = c.All(ctx, &posts)
	if err != nil {
		return nil, fmt.Errorf("mongodb get all posts: %w", err)
	}

	return posts, nil
}

func (p *PostMongoDB) CreatePost(category, title, typePost, url, text string, author *user.User) (*posts.Post, error) {
	newPost := &posts.Post{
		Author:           author,
		Category:         category,
		Comments:         []*posts.Comment{},
		Created:          time.Now(),
		ID:               uuid.New().String(),
		Score:            InitialScore,
		Title:            title,
		Type:             typePost,
		Text:             text,
		UpvotePercentage: FullPercent,
		URL:              url,
		Views:            ZeroViews,
		Votes: []*posts.Vote{
			{
				UserID: author.ID,
				Vote:   LIKE,
			},
		},
	}

	ctx, cancel := p.withTimeout()
	defer cancel()
	if _, err := p.db.InsertOne(ctx, newPost); err != nil {
		return nil, fmt.Errorf("mongodb create post: %w", err)
	}
	return newPost, nil
}

func (p *PostMongoDB) GetPost(postID string) (*posts.Post, error) {
	var post *posts.Post
	filter := bson.M{"_id": postID}
	ctx, cancel := p.withTimeout()
	defer cancel()
	err := p.db.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("mongodb get post: %w", myerrors.ErrNoPost)
		}
		return nil, fmt.Errorf("mongodb get post: %w", err)
	}
	return post, nil
}

func (p *PostMongoDB) DeletePost(postID, userID string) error {
	post, err := p.GetPost(postID)
	if err != nil {
		return fmt.Errorf("mongodb delete post: %w", err)
	}

	if post.Author.ID != userID {
		return fmt.Errorf("mongodb delete post: %s %s", NoCredentialsToDelete, postID)
	}

	filter := bson.M{"_id": postID}
	ctx, cancel := p.withTimeout()
	defer cancel()
	if res, err := p.db.DeleteOne(ctx, filter); err != nil || res.DeletedCount == 0 {
		return fmt.Errorf("mongodb delete post: %w", err)
	}

	return nil
}

func (p *PostMongoDB) GetPostsByCategory(category string) ([]*posts.Post, error) {
	var posts []*posts.Post
	filter := bson.M{"category": category}
	ctx, cancel := p.withTimeout()
	defer cancel()
	c, err := p.db.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mogngodb get posts by category: %w", err)
	}
	err = c.All(ctx, &posts)
	if err != nil {
		return nil, fmt.Errorf("mogngodb get posts by category: %w", err)
	}

	return posts, nil
}

func (p *PostMongoDB) CreateComment(postID, text string, author *user.User) (*posts.Post, error) {
	comment := &posts.Comment{
		Author:  author,
		Body:    text,
		Created: time.Now(),
		ID:      uuid.New().String(),
	}

	filterPost := bson.M{"_id": postID}
	update := bson.M{
		"$push": bson.M{"comments": comment},
	}

	ctx, cancel := p.withTimeout()
	defer cancel()
	if _, err := p.db.UpdateOne(ctx, filterPost, update); err != nil {
		return nil, fmt.Errorf("mogngodb create comment: %w", err)
	}

	post, err := p.GetPost(postID)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (p *PostMongoDB) DeleteComment(postID string, commentID string) (*posts.Post, error) {
	filterPost := bson.M{"_id": postID}
	update := bson.M{
		"$pull": bson.M{"comments": bson.M{"_id": commentID}},
	}

	ctx, cancel := p.withTimeout()
	defer cancel()
	if _, err := p.db.UpdateOne(ctx, filterPost, update); err != nil {
		return nil, fmt.Errorf("mogngodb delete comment: %w", err)
	}

	post, err := p.GetPost(postID)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (p *PostMongoDB) getUpvotePercentage(post *posts.Post) int {
	if len(post.Votes) == 0 {
		return ZeroPercent
	}

	voteCount := 0
	for _, vote := range post.Votes {
		if vote.Vote == LIKE {
			voteCount++
		}
	}

	return voteCount * FullPercent / len(post.Votes)
}

func (p *PostMongoDB) updateUpvotePercentage(postID string) error {
	post, err := p.GetPost(postID)
	if err != nil {
		return fmt.Errorf("mogngodb update upvote percentage: %w", err)
	}
	newUpvotePercentage := p.getUpvotePercentage(post)
	update := bson.M{
		"$set": bson.M{
			"upvotePercentage": newUpvotePercentage,
		},
	}

	filter := bson.M{"_id": postID}
	ctx, cancel := p.withTimeout()
	defer cancel()
	if _, err = p.db.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("mogngodb update upvote percentage: %w", err)
	}
	return nil
}

func (p *PostMongoDB) getScore(post *posts.Post) int {
	score := 0
	for _, vote := range post.Votes {
		score += vote.Vote
	}
	return score
}

func (p *PostMongoDB) updateScore(postID string) error {
	post, err := p.GetPost(postID)
	if err != nil {
		return fmt.Errorf("mogngodb update score: %w", err)
	}
	newScore := p.getScore(post)
	update := bson.M{
		"$set": bson.M{
			"score": newScore,
		},
	}

	filter := bson.M{"_id": postID}
	ctx, cancel := p.withTimeout()
	defer cancel()
	if _, err = p.db.UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("mogngodb update score: %w", err)
	}
	return nil
}

func (p *PostMongoDB) updateMetrics(postID string) error {
	if err := p.updateScore(postID); err != nil {
		return fmt.Errorf("mogngodb update metrics: %w", err)
	}
	if err := p.updateUpvotePercentage(postID); err != nil {
		return fmt.Errorf("mogngodb update metrics: %w", err)
	}
	return nil
}

func (p *PostMongoDB) vote(postID string, userID string, vote int) error {
	filter := bson.M{"_id": postID, "votes.user": userID}
	update := bson.M{
		"$set": bson.M{
			"votes.$.vote": vote,
		},
	}
	ctx, cancel := p.withTimeout()
	defer cancel()

	result, err := p.db.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("mogngodb vote: %w", err)
	}

	if result.ModifiedCount == 0 {
		filter := bson.M{"_id": postID}
		update := bson.M{
			"$push": bson.M{
				"votes": bson.M{
					"user": userID,
					"vote": vote,
				},
			},
		}
		if _, err = p.db.UpdateOne(ctx, filter, update); err != nil {
			return fmt.Errorf("mogngodb vote: %w", err)
		}
	}

	if err = p.updateMetrics(postID); err != nil {
		return fmt.Errorf("mogngodb vote: %w", err)
	}

	return nil
}

func (p *PostMongoDB) UpvotePost(postID, userID string) (*posts.Post, error) {
	err := p.vote(postID, userID, LIKE)
	if err != nil {
		return nil, fmt.Errorf("mogngodb upvote post: %w", err)
	}

	post, err := p.GetPost(postID)
	if err != nil {
		return nil, fmt.Errorf("mogngodb upvote post: %w", err)
	}

	return post, nil
}

func (p *PostMongoDB) UnvotePost(postID, userID string) (*posts.Post, error) {
	filter := bson.M{"_id": postID}
	update := bson.M{
		"$pull": bson.M{
			"votes": bson.M{
				"user": userID,
			},
		},
	}
	ctx, cancel := p.withTimeout()
	defer cancel()
	if _, err := p.db.UpdateOne(ctx, filter, update); err != nil {
		return nil, fmt.Errorf("mogngodb unvote post: %w", err)
	}

	if err := p.updateMetrics(postID); err != nil {
		return nil, fmt.Errorf("mogngodb vote: %w", err)
	}

	post, err := p.GetPost(postID)
	if err != nil {
		return nil, fmt.Errorf("mogngodb unvote post: %w", err)
	}

	return post, nil
}

func (p *PostMongoDB) DownvotePost(postID, userID string) (*posts.Post, error) {
	err := p.vote(postID, userID, DISLIKE)
	if err != nil {
		return nil, fmt.Errorf("mogngodb downvote post: %w", err)
	}

	var post *posts.Post
	filter := bson.M{"_id": postID}
	ctx, cancel := p.withTimeout()
	defer cancel()
	err = p.db.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		return nil, fmt.Errorf("mogngodb downvote post: %w", err)
	}

	return post, nil
}

func (p *PostMongoDB) GetPostsByUser(username string) ([]*posts.Post, error) {
	var posts []*posts.Post
	filter := bson.M{"author.username": username}
	ctx, cancel := p.withTimeout()
	defer cancel()
	c, err := p.db.Find(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("mogngodb get posts by id: %w", err)
	}
	err = c.All(ctx, &posts)
	if err != nil {
		return nil, fmt.Errorf("mogngodb get posts by id: %w", err)
	}

	return posts, nil
}
