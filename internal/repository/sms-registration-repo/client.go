package smsregistrationrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

var ErrSMSSendingFailed = errors.New("sms sending failed")

type SMSService struct {
	cfg Config
}

func New(cfg Config) *SMSService {
	return &SMSService{
		cfg: cfg,
	}
}

func (s *SMSService) SendOTP(ctx context.Context, phone string, otp string) error {
	message := "Your verification code: " + otp

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("https://%s:%s@gate.smsaero.ru/v2/sms/send?number=%s&text=%s&sign=%s", s.cfg.Email, s.cfg.ApiKey, phone, message, s.cfg.Sender),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("failed to send SMS: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%w with status: %s", ErrSMSSendingFailed, resp.Status)
	}

	var response struct {
		Success bool `json:"success"`
		Data    []struct {
			Id           int       `json:"id"`
			From         string    `json:"from"`
			Number       string    `json:"number"`
			Text         string    `json:"text"`
			Status       int       `json:"status"`
			ExtendStatus string    `json:"extendStatus"`
			Channel      string    `json:"channel"`
			Cost         float64   `json:"cost"`
			DateCreated  time.Time `json:"dateCreated"`
			DateSend     time.Time `json:"dateSend"`
		} `json:"data"`
		Message string `json:"message"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("%w: %v", ErrSMSSendingFailed, response)
	}

	return nil
}
