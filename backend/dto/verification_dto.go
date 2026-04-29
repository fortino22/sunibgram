package dto

type SendVerificationRequest struct {
	Email string `json:"email"`
	Type  string `json:"type"`
}

type VerifyCodeRequest struct {
	Email string `json:"email"`
	Code  string `json:"code"`
	Type  string `json:"type"`
}
