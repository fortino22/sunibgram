package service

type IVerificationService interface {
	SendCode(email string) error
	VerifyCode(email, code string) error
}
