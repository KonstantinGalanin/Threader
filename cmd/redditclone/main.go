package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
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
	err := godotenv.Load()

	redisURL := "redis://"
	if redisUser := os.Getenv("REDIS_USER"); redisUser != "" {
		redisURL += redisUser + ":@"
	}
	redisURL += "localhost:6379/0"
	redisConn, err := redis.DialURL(redisURL)

	if err != nil {
		logrus.WithError(err).Fatal("Conn redis error")
	}
	redisManager := sessionRepository.NewSessionManagerRedis(redisConn)


	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName + "?"
	dsn += "&charset=utf8&interpolateParams=true"
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
	sessMongo, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("MONGO_URI")))
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
