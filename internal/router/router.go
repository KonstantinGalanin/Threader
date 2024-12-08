package router

import (
	"html/template"
	"net/http"

	"github.com/KonstantinGalanin/redditclone/internal/middleware"
	"github.com/KonstantinGalanin/redditclone/internal/session"
	"github.com/gorilla/mux"

	postsHandlers "github.com/KonstantinGalanin/redditclone/internal/posts/handlers"
	userHandlers "github.com/KonstantinGalanin/redditclone/internal/user/handlers"
)

func renderStatic(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("static/html/index.html"))
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func NewRouter(
	userHandler userHandlers.UserHandler,
	postsHandler postsHandlers.PostsHandler,
	sessionManager session.SessionManager,

) http.Handler {
	publicRouter := mux.NewRouter()
	privateRouter := publicRouter.NewRoute().Subrouter()

	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	publicRouter.PathPrefix("/static/").Handler(staticHandler)

	publicRouter.HandleFunc("/api/register", userHandler.Signup).Methods(http.MethodPost)
	publicRouter.HandleFunc("/api/login", userHandler.Login).Methods(http.MethodPost)

	privateRouter.HandleFunc("/api/posts", postsHandler.CreatePost).Methods(http.MethodPost)
	publicRouter.HandleFunc("/api/posts/", postsHandler.GetAll).Methods(http.MethodGet)
	publicRouter.HandleFunc("/api/posts/{category}", postsHandler.GetByCategory).Methods(http.MethodGet)

	publicRouter.HandleFunc("/api/post/{id}", postsHandler.GetPost).Methods(http.MethodGet)
	privateRouter.HandleFunc("/api/post/{id}", postsHandler.CreateComment).Methods(http.MethodPost)
	privateRouter.HandleFunc("/api/post/{id}", postsHandler.DeletePost).Methods(http.MethodDelete)
	privateRouter.HandleFunc("/api/post/{id}/{commentID}", postsHandler.DeleteComment).Methods(http.MethodDelete)
	privateRouter.HandleFunc("/api/post/{id}/upvote", postsHandler.UpvotePost).Methods(http.MethodGet)
	privateRouter.HandleFunc("/api/post/{id}/unvote", postsHandler.UnvotePost).Methods(http.MethodGet)
	privateRouter.HandleFunc("/api/post/{id}/downvote", postsHandler.DownvotePost).Methods(http.MethodGet)

	publicRouter.HandleFunc("/api/user/{username}", postsHandler.PostsByUser).Methods(http.MethodGet)

	publicRouter.PathPrefix("/").HandlerFunc(renderStatic).Methods(http.MethodGet)

	publicRouter.Use(middleware.AccessLog)
	privateRouter.Use(middleware.Session(sessionManager))
	publicRouter.Use(middleware.Panic)

	return publicRouter
}
