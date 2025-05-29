package email

const (
	EmailTypeRegister       = "REGISTER"
	EmailTypeForgetPassword = "FORGET_PASSWORD"
	EmailTypeResetPassword  = "RESET_PASSWORD"
)

type EmailTemplate struct {
	Subject        string
	TemplateFile   string
	RequiredFields []string
}

var EmailTemplates = map[string]EmailTemplate{
	EmailTypeRegister: {
		Subject:        "Welcome to Our Service - Complete Registration",
		TemplateFile:   "register.html",
		RequiredFields: []string{"name", "otp"},
	},
	EmailTypeForgetPassword: {
		Subject:        "Password Reset Request",
		TemplateFile:   "forget_password.html",
		RequiredFields: []string{"name", "otp"},
	},
	EmailTypeResetPassword: {
		Subject:        "Your Password Has Been Reset",
		TemplateFile:   "reset_password.html",
		RequiredFields: []string{"name"},
	},
}
