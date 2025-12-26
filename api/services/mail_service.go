package services

import (
	"fmt"

	"github.com/resend/resend-go/v2"
	"github.com/seojoonrp/bapddang-server/config"
)

func (s *userService) sendVerificationEmail(email string, token string) error {
	resendClient := resend.NewClient(config.AppConfig.ResendAPIKey)

	verifyURL := fmt.Sprintf("https://api.sslip.io/auth/verify-email?token=%s", token)

	params := &resend.SendEmailRequest{
		From: "Bobttaeng <noreply@happetite.bobttaeng.com>",
		To: []string{email},
		Subject: "[밥땡] 이메일 인증을 완료해주세요",
		Html: fmt.Sprintf("<p>가입을 축하합니다! <a href='%s'>여기</a>를 눌러 인증을 완료하세요.</p>", verifyURL),
	}

	_, err := resendClient.Emails.Send(params)
	if err != nil {
		fmt.Println("Failed to send verification email:", err)
		return err
	}
	return nil
}