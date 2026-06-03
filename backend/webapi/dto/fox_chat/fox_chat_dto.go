package foxchatdto

// QAPairRequest is one prior turn the client replays for short-term context.
type QAPairRequest struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// AskRequest is the body for POST /fox/ask. SessionID is 0 to start a new chat
// session; the response carries the resolved id in the X-Fox-Session-Id header.
type AskRequest struct {
	SessionID uint            `json:"session_id"`
	Question  string          `json:"question"`
	History   []QAPairRequest `json:"history"`
}
