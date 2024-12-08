package user

type User struct {
	Username string `json:"username" bson:"username"`
	Password string `bson:"password"`
	ID       string `json:"id" bson:"_id"`
}

//go:generate mockgen -source=user.go -destination=repository/repo_mock.go -package=repository ItemRepo
type UserRepo interface {
	Signup(username, password string) (*User, error)
	Login(username, password string) (*User, error)
	GetUserByUsername(username string) (*User, error)
}
