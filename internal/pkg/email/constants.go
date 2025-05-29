package email

const (
	EmailTypeRegister       = "REGISTER"
	EmailTypeForgetPassword = "FORGET_PASSWORD"
	EmailTypeResetPassword  = "RESET_PASSWORD"
)

type EmailTemplate struct {
	Subject      string
	TemplateFile string
}

var EmailTemplates = map[string]EmailTemplate{
	EmailTypeRegister: {
		Subject:      "Welcome to Quick Cart - Complete Your Registration",
		TemplateFile: "register.html",
	},
	EmailTypeForgetPassword: {
		Subject:      "Quick Cart - Password Reset Request",
		TemplateFile: "forget_password.html",
	},
	EmailTypeResetPassword: {
		Subject:      "Your Quick Cart Password Has Been Reset",
		TemplateFile: "reset_password.html",
	},
}
