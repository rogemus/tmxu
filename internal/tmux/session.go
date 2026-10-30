package tmux

type Session struct {
	Id      string   `json:"id"`
	Name    string   `json:"name"`
	Windows []Window `json:"windows"`
}

func NewSession(l TmuxLine) Session {
	return Session{
		Id:      l.SessionID,
		Name:    l.SessionName,
		Windows: make([]Window, 0),
	}
}

func (s Session) Eq(c Comperable) bool {
	i, ok := c.(Session)
	return ok && s.Id == i.Id
}
