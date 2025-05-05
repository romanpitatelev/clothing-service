package smsregistrationrepo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"
)

var ErrSMSSendingFailed = errors.New("sms sending failed")

type Config struct {
	Host     string
	Schema   string
	Email    string
	ApiKey   string
	Sender   string
	TestMode bool
}

type SMSService struct {
	cfg Config
}

func New(cfg Config) *SMSService {
	return &SMSService{
		cfg: cfg,
	}
}

type Response struct {
	Success bool `json:"success"`
	Data    struct {
		Id           int     `json:"id"`
		From         string  `json:"from"`
		Number       string  `json:"number"`
		Text         string  `json:"text"`
		Status       int     `json:"status"`
		ExtendStatus string  `json:"extendStatus"`
		Channel      string  `json:"channel"`
		Cost         float64 `json:"cost"`
		DateCreated  string  `json:"dateCreated"`
		DateSend     int     `json:"dateSend"`
	} `json:"data"`
	Message string `json:"message"`
}

func (s *SMSService) SendOTP(ctx context.Context, phone string, otp string) error {
	data := url.Values{}
	data.Set("number", phone)
	data.Set("sign", s.cfg.Sender)
	data.Set("text", EnrichOTP(otp))

	endpoint := "send"
	if s.cfg.TestMode {
		endpoint = "testsend"
	}

	smsURL := url.URL{
		Scheme:   s.cfg.Schema,
		Host:     s.cfg.Host,
		Path:     "/v2/sms/" + endpoint,
		RawQuery: data.Encode(),
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		smsURL.String(),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	request.SetBasicAuth(s.cfg.Email, s.cfg.ApiKey)

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

	var response Response

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("%w: %v", ErrSMSSendingFailed, response)
	}

	return nil
}

func EnrichOTP(otp string) string {
	return "Код верификации: " + otp
}

func ExtractOTP(val string) string {
	items := strings.Split(val, ": ")
	if len(items) < 2 { //nolint:mnd
		return ""
	}

	return items[1]
}
