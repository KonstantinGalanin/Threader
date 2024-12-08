package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"

	"github.com/KonstantinGalanin/redditclone/internal/posts"
	"github.com/KonstantinGalanin/redditclone/internal/user"
)

var mockCollection *mongo.Collection

func TestNewPostMongoDB(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test", func(mt *mtest.T) {
		mockCollection = mt.Coll
		assert.NotNil(t, mockCollection)
		mockDB := NewPostMongoDB(mockCollection)

		assert.NotNil(t, mockDB)
	})
}

func TestWithTimeout(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mockCollection := mt.Coll
	mockDB := NewPostMongoDB(mockCollection)

	ctx, cancel := mockDB.withTimeout()

	assert.NotNil(t, mockDB)
	assert.NotNil(t, ctx)
	assert.NotNil(t, cancel)
}

func TestCreatePost(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	cases := []struct {
		name        string
		resp        primitive.D
		expectError bool
	}{
		{
			name:        "success",
			resp:        mtest.CreateSuccessResponse(),
			expectError: false,
		},
		{
			name:        "insert error",
			resp:        mtest.CreateWriteErrorsResponse(mtest.WriteError{Index: 1, Code: 11000, Message: "duplicate key error"}),
			expectError: true,
		},
	}

	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)

			mt.AddMockResponses(c.resp)

			posts, err := mockDB.CreatePost("category", "title", "type", "url", "tet", &user.User{
				Username: "username",
				Password: "password",
				ID:       "1",
			})
			if c.expectError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestGetAllPosts(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	cases := []struct {
		name          string
		resp          []bson.D
		expectedError bool
	}{
		{
			name: "success",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, bson.D{}),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: false,
		},
		{
			name: "collection error",
			resp: []bson.D{
				mtest.CreateWriteErrorsResponse(),
			},
			expectedError: true,
		},
		{
			name: "cursor error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, bson.D{{Key: "author", Value: 1}}),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: true,
		},
	}

	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockDB := NewPostMongoDB(mt.Coll)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.GetAllPosts()
			if c.expectedError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestGetPost(t *testing.T) {
	cases := []struct {
		name          string
		resp          []bson.D
		expectedError bool
	}{
		{
			name: "success",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, bson.D{}),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: false,
		},
		{
			name: "no documents",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: true,
		},
		{
			name: "no documents",
			resp: []bson.D{
				mtest.CreateWriteErrorsResponse(),
			},
			expectedError: true,
		},
	}

	postID := "1"
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockDB := NewPostMongoDB(mt.Coll)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.GetPost(postID)
			if c.expectedError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

var postData bson.D = bson.D{
	{Key: "_id", Value: "1"},
	{Key: "title", Value: "title"},
	{Key: "category", Value: "Music"},
	{Key: "score", Value: 10},
	{Key: "type", Value: "text"},
	{Key: "upvotePercentage", Value: 85},
	{Key: "url", Value: "https://url.com"},
	{Key: "text", Value: "Text"},
	{Key: "views", Value: 100},
	{Key: "created", Value: time.Now()},
	{Key: "author", Value: bson.D{
		{Key: "username", Value: "User"},
		{Key: "_id", Value: "2"},
		{Key: "password", Value: "password"},
	}},
	{Key: "comments", Value: bson.A{
		bson.D{{Key: "_id", Value: "1"}},
		bson.D{{Key: "created", Value: time.Now()}},
		bson.D{{Key: "body", Value: "comment"}},
		bson.D{{Key: "author", Value: bson.D{
			{Key: "username", Value: "User"},
			{Key: "_id", Value: "1"},
			{Key: "password", Value: "password"},
		}}},
	}},
	{Key: "votes", Value: bson.A{
		bson.D{{Key: "user", Value: "User"}, {Key: "vote", Value: 1}},
	}},
}

func TestDeletePost(t *testing.T) {
	cases := []struct {
		name          string
		resp          []bson.D
		postID        string
		userID        string
		expectedError bool
	}{
		{
			postID:        "1",
			userID:        "2",
			name:          "get post error",
			resp:          nil,
			expectedError: true,
		},
		{
			postID: "invalid post id",
			userID: "invalid user id",
			name:   "get post error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectedError: true,
		},
		{
			postID: "1",
			userID: "2",
			name:   "delete error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 0}},
			},
			expectedError: true,
		},
		{
			postID: "1",
			userID: "2",
			name:   "success",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectedError: false,
		},
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockDB := NewPostMongoDB(mt.Coll)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			err := mockDB.DeletePost(c.postID, c.userID)
			if c.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPostsByCategory(t *testing.T) {
	cases := []struct {
		name          string
		resp          []bson.D
		category      string
		expectedError bool
	}{
		{
			name:          "not found",
			category:      "invalid category",
			resp:          nil,
			expectedError: true,
		},
		{
			name:     "cursor error",
			category: "Music",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, bson.D{{Key: "author", Value: 1}}),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: true,
		},
		{
			name:     "success",
			category: "Music",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, bson.D{}),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: false,
		},
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockDB := NewPostMongoDB(mt.Coll)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.GetPostsByCategory(c.category)
			if c.expectedError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestCreateComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := "1"
	text := "text"
	author := &user.User{
		Username: "User",
		Password: "password",
		ID:       "1",
	}

	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name:        "upddate error",
			resp:        nil,
			expectError: true,
		},
		{
			name: "get post error 2",
			resp: []bson.D{
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "success",
			resp: []bson.D{
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.CreateComment(postID, text, author)

			if c.expectError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestDeleteComment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := "1"
	commentID := "1"

	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name:        "upddate error",
			resp:        nil,
			expectError: true,
		},
		{
			name: "get post error 2",
			resp: []bson.D{
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "success",
			resp: []bson.D{
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.DeleteComment(postID, commentID)

			if c.expectError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestGetUpvotePercentage(t *testing.T) {
	cases := []struct {
		name     string
		post     *posts.Post
		returned int
	}{
		{
			name:     "empty votes",
			post:     &posts.Post{},
			returned: 0,
		},
		{
			name: "fifty-fifty",
			post: &posts.Post{
				Votes: []*posts.Vote{
					&posts.Vote{
						UserID: "1",
						Vote:   1,
					},
					&posts.Vote{
						UserID: "2",
						Vote:   -1,
					},
				},
			},
			returned: 50,
		},
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockDB := NewPostMongoDB(mt.Coll)
			res := mockDB.getUpvotePercentage(c.post)
			assert.Equal(t, res, c.returned)
		})
	}
}

func TestUpdateUpvotePercentage(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := "1"

	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name:        "get post error",
			resp:        nil,
			expectError: true,
		},
		{
			name: "upddate error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: true,
		},
		{
			name: "success",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			err := mockDB.updateUpvotePercentage(postID)

			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetScore(t *testing.T) {
	post := &posts.Post{
		Votes: []*posts.Vote{
			&posts.Vote{UserID: "1", Vote: 1},
			&posts.Vote{UserID: "2", Vote: 1},
		},
	}
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("success", func(mt *mtest.T) {
		mockCollection := mt.Coll
		mockDB := NewPostMongoDB(mockCollection)
		res := mockDB.getScore(post)
		assert.Equal(t, 2, res)
	})
}

func TestUpdateScore(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := "1"

	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name:        "get post error",
			resp:        nil,
			expectError: true,
		},
		{
			name: "upddate error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: true,
		},
		{
			name: "success",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			err := mockDB.updateScore(postID)

			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

var voteSuccess = []bson.D{
	mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
	mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
	mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
	mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
	{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
	mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
	mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
	{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
	mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
	mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
}

func TestVote(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	postID := "1"
	userID := "1"
	vote := 1

	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name:        "upddate error",
			resp:        nil,
			expectError: true,
		},
		{
			name: "upddate error 2",
			resp: []bson.D{
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: true,
		},
		{
			name: "update metrics error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: true,
		},
		{
			name:        "success",
			resp:        voteSuccess,
			expectError: false,
		},
	}

	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			err := mockDB.vote(postID, userID, vote)

			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUpdateMetrics(t *testing.T) {
	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name: "updateUpvotePercentage error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "success",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	postID := "1"
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			err := mockDB.updateMetrics(postID)

			if c.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}

}

func TestUpvotePost(t *testing.T) {
	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name: "vote error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "get post error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "get post error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	postID := "1"
	userID := "1"
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.UpvotePost(postID, userID)

			if c.expectError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestDownvotePost(t *testing.T) {
	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name: "vote error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "get post error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "get post error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	postID := "1"
	userID := "1"
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.DownvotePost(postID, userID)

			if c.expectError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestGetPostsByUser(t *testing.T) {
	cases := []struct {
		name          string
		resp          []bson.D
		expectedError bool
	}{
		{
			name:          "not found",
			resp:          nil,
			expectedError: true,
		},
		{
			name: "cursor error",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, bson.D{{Key: "author", Value: 1}}),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: true,
		},
		{
			name: "success",
			resp: []bson.D{
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, bson.D{}),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectedError: false,
		},
	}

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	username := "User"
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockDB := NewPostMongoDB(mt.Coll)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.GetPostsByUser(username)
			if c.expectedError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}

func TestUnvotePost(t *testing.T) {
	cases := []struct {
		name        string
		resp        []bson.D
		expectError bool
	}{
		{
			name:        "update error",
			resp:        nil,
			expectError: true,
		},
		{
			name: "update metrics error",
			resp: []bson.D{
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "get post 2",
			resp: []bson.D{
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
			},
			expectError: true,
		},
		{
			name: "success",
			resp: []bson.D{
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
				{{Key: "ok", Value: 1}, {Key: "acknowledged", Value: true}, {Key: "n", Value: 1}},
				mtest.CreateCursorResponse(1, "posts.post", mtest.FirstBatch, postData),
				mtest.CreateCursorResponse(0, "posts.post", mtest.NextBatch),
			},
			expectError: false,
		},
	}

	postID := "1"
	userID := "1"
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	for _, c := range cases {
		mt.Run(c.name, func(mt *mtest.T) {
			mockCollection := mt.Coll
			mockDB := NewPostMongoDB(mockCollection)
			for _, response := range c.resp {
				mt.AddMockResponses(response)
			}
			posts, err := mockDB.UnvotePost(postID, userID)

			if c.expectError {
				assert.Error(t, err)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, posts)
			}
		})
	}
}
