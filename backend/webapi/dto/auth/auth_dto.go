package authdto

type RegisterRequest struct {
	Name            string `json:"name"`
	UserName        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	Gender          string `json:"gender" validate:"required"`
	BirthdayDate    string `json:"birthday_date"`
	TermsAccepted   bool   `json:"terms_accepted"`
}

type ChangePasswordRequest struct {
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type ForgotPasswordRequest struct {
	Email           string `json:"email"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message     string `json:"message"`
	AccessToken string `json:"access_token"`
	SID         string `json:"sid"`
}

type RefreshRequest struct {
	SID string `json:"sid"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	SID         string `json:"sid"`
}

type UserData struct {
	UserID   uint   `json:"user_id"`
	Name     string `json:"name"`
	UserName string `json:"username"`
	Email    string `json:"email"`
	RoleID   uint   `json:"role_id"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
