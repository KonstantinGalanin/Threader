package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	postsHandlers "github.com/KonstantinGalanin/redditclone/internal/posts/handlers"
	postsRepository "github.com/KonstantinGalanin/redditclone/internal/posts/repository"
	"github.com/KonstantinGalanin/redditclone/internal/router"
	sessionRepository "github.com/KonstantinGalanin/redditclone/internal/session/redis"
	"github.com/KonstantinGalanin/redditclone/internal/token_manager/jwt"
	userHandlers "github.com/KonstantinGalanin/redditclone/internal/user/handlers"
	userRepository "github.com/KonstantinGalanin/redditclone/internal/user/repository"
)

var (
	ADDR = os.Getenv("ADDR")
)

func main() {
	redisConn, err := redis.DialURL("redis://user:@localhost:6379/0")
	if err != nil {
		logrus.WithError(err).Fatal("Conn redis error")
	}
	redisManager := sessionRepository.NewSessionManagerRedis(redisConn)

	dsn := "root:love@tcp(localhost:3306)/users?"
	dsn += "&charset=utf8"
	dsn += "&interpolateParams=true"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		logrus.WithError(err).Fatal("Open mysql error")
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		logrus.WithError(err).Fatal("Connect mysql error")
	}

	userHandler := userHandlers.UserHandler{
		UserRepo:       userRepository.NewUserPostgresRepo(db),
		SessionManager: redisManager,
		JwtService:     jwt.NewJwtService(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	sessMongo, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost"))
	if err != nil {
		logrus.WithError(err).Fatal("Open mongodb error")
	}
	collection := sessMongo.Database("reddit").Collection("posts")

	postsHandler := postsHandlers.PostsHandler{
		PostsRepo:      postsRepository.NewPostMongoDB(collection),
		UserRepo:       userHandler.UserRepo,
		SessionManager: redisManager,
	}

	r := router.NewRouter(userHandler, postsHandler, redisManager)

	logrus.SetFormatter(&logrus.TextFormatter{DisableColors: true})
	logrus.WithFields(logrus.Fields{
		"type": "START",
	}).Info("starting server")

	addr := ADDR
	if addr == "" {
		addr = ":8000"
	}
	err = http.ListenAndServe(addr, r)
	if err != nil {
		logrus.WithError(err).Fatal("Starting server error")
	}
}
