package repository

var (
	CheckExists = "SELECT EXISTS(SELECT 1 FROM users WHERE username = ?);"
	CreateUser  = "INSERT INTO users (id, username, password) VALUES (?, ?, ?);"
	GetUser     = "SELECT id, username, password FROM users WHERE username = ?;"
)
