package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"rest-api/configs"
	"rest-api/utils"
)

type EmailService struct {
	from     string
	password string
	smtpHost string
	smtpPort string
}

func NewEmailService() *EmailService {
	return &EmailService{
		from:     configs.AppEnv.SMTPEmail,
		password: configs.AppEnv.SMTPPassword,
		smtpHost: "smtp.gmail.com",
		smtpPort: "587",
	}
}

func (e *EmailService) SendOTP(email, otp string) error {
	subject := "Password Reset OTP"
	body := fmt.Sprintf(`<html>
<body style="font-family: Arial, sans-serif; padding: 20px;">
    <div style="max-width: 500px; margin: 0 auto; background: #f9f9f9; padding: 30px; border-radius: 10px;">
        <h2 style="color: #333; text-align: center;">Password Reset Request</h2>
        <p style="color: #666; font-size: 16px;">Your OTP for password reset is:</p>
        <div style="background: #007bff; color: white; font-size: 32px; font-weight: bold; 
            padding: 15px; text-align: center; border-radius: 5px; letter-spacing: 5px; margin: 20px 0;">
            %s
        </div>
        <p style="color: #666; font-size: 14px;">This OTP will expire in <strong>5 minutes</strong>.</p>
        <p style="color: #999; font-size: 12px; margin-top: 30px;">
            If you didn't request this, please ignore this email. 
            For security reasons, never share your OTP with anyone.
        </p>
    </div>
</body>
</html>`, otp)

	return e.sendEmail(email, subject, body)
}

func (e *EmailService) sendEmail(to, subject, body string) error {
	headers := make(map[string]string)
	headers["From"] = e.from
	headers["To"] = to
	headers["Subject"] = subject
	headers["MIME-Version"] = "1.0"
	headers["Content-Type"] = "text/html; charset=utf-8"

	var message strings.Builder
	for k, v := range headers {
		message.WriteString(fmt.Sprintf("%s: %s\r\n", k, v))
	}
	message.WriteString("\r\n")
	message.WriteString(body)

	addr := fmt.Sprintf("%s:%s", e.smtpHost, e.smtpPort)

	auth := smtp.PlainAuth("", e.from, e.password, e.smtpHost)

	err := smtp.SendMail(addr, auth, e.from, []string{to}, []byte(message.String()))
	if err != nil {
		utils.LogErrorPlain("Failed to send email to " + to + ": " + err.Error())
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func GenerateSecureOTP(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	encoded := base64.URLEncoding.EncodeToString(bytes)
	otp := strings.ReplaceAll(encoded, "+", "0")
	otp = strings.ReplaceAll(otp, "/", "1")
	otp = strings.ReplaceAll(otp, "=", "2")

	if len(otp) < length {
		return "", fmt.Errorf("failed to generate OTP of length %d", length)
	}

	return otp[:length], nil
}

func HashOTP(otp string) string {
	return utils.HashToken(otp)
}

func GetOTPExpiry() time.Time {
	return time.Now().Add(5 * time.Minute)
}
