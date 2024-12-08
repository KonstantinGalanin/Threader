package session

type Session struct {
	Username string `json:"username"`
}

type SessionID struct {
	ID string
}

//go:generate mockgen -source=session.go -destination=mock/session_mock.go -package=mock Session
type SessionManager interface {
	Create(in *Session) (*SessionID, error)
	Check(in *SessionID) (*Session, error)
	Delete(in *SessionID) error
}
