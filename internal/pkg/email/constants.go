package email

const (
	EmailTypeRegister       = "USER_REGISTER"
	EmailTypeForgetPassword = "FORGET_PASSWORD"
	EmailTypeResetPassword  = "RESET_PASSWORD"
)

type EmailPayload struct {
	To         string
	EMAIL_TYPE string
	Data       map[string]interface{}
}

type EmailTemplate struct {
	Subject        string
	TemplateFile   string
	RequiredFields []string
}

var EmailTemplates = map[string]EmailTemplate{
	EmailTypeRegister: {
		Subject:        "Welcome to Ride Sharing Service - Complete Registration",
		TemplateFile:   "internal/pkg/email/templates/register.html",
		RequiredFields: []string{"name", "otp"},
	},
	EmailTypeForgetPassword: {
		Subject:        "Password Reset Request",
		TemplateFile:   "internal/pkg/email/templates/forget_password.html",
		RequiredFields: []string{"name", "otp"},
	},
	EmailTypeResetPassword: {
		Subject:        "Your Password Has Been Reset",
		TemplateFile:   "internal/pkg/email/templates/reset_password.html",
		RequiredFields: []string{"name"},
	},
}
