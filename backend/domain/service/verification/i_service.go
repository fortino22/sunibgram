package verificationsvciface

import verificationcodemodel "github.com/sunibgram/domain/model/verification_code"

type IVerificationService interface {
	SendCode(email string, t verificationcodemodel.VerificationType) error
	VerifyCode(email, code string, t verificationcodemodel.VerificationType) error
}
