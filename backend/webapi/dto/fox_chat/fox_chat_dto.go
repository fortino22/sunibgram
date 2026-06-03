package foxchatdto

// QAPairRequest is one prior turn the client replays for short-term context.
type QAPairRequest struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// AskRequest is the body for POST /fox/ask.
type AskRequest struct {
	Question string          `json:"question"`
	History  []QAPairRequest `json:"history"`
}
