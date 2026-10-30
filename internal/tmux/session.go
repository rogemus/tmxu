package tmux

type Session struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	FirstWindowName string `json:"firstWindowName"`
}

func NewSession(l TmuxLine) Session {
	return Session{
		ID:              l.SessionID,
		Name:            l.SessionName,
		FirstWindowName: l.WindowName,
	}
}

func (s Session) Eq(c Comperable) bool {
	i, ok := c.(Session)
	return ok && s.ID == i.ID
}
