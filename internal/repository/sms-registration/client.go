package smsregistration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

type SMSService struct {
	cfg          Config
	verification map[string]verificationData
	mu           sync.RWMutex
}

type verificationData struct {
	Code      string
	ExpiresAt time.Time
}

func New(cfg Config) *SMSService {
	return &SMSService{
		cfg:          cfg,
		verification: make(map[string]verificationData),
	}
}

func (s *SMSService) StartCleanup() {
	ticker := time.NewTicker(s.cfg.CleanupPeriod)
	defer ticker.Stop()

	go func() {
		for range ticker.C {
			s.cleanupExpiredCodes()
		}
	}()
}

func (s *SMSService) cleanupExpiredCodes() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	for phone, data := range s.verification {
		if now.After(data.ExpiresAt) {
			delete(s.verification, phone)
		}
	}
}

func (s *SMSService) SendVerificationCode(ctx context.Context, phone string) (*VerificationResponse, error) {
	code := s.generateCode()
	expiresAt := time.Now().Add(s.cfg.CodeValidityDuration)

	s.mu.Lock()
	s.verification[phone] = verificationData{
		Code:      code,
		ExpiresAt: expiresAt,
	}

	message := fmt.Sprintf("Your verification code: %s", code)

	requestBody := []map[string]interface{}{
		{
			"channelType":       "SMS",
			"senderName":        s.cfg.SenderName,
			"destination":       phone,
			"content":           message,
			"externalMessageId": code,
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/api/v1/message", s.cfg.BaseURL),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	request.Header.Set("Authorization", fmt.Sprintf("Basic %s", s.cfg.AuthToken))
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send SMS: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Warn().Err(err).Msg("failed to close response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("SMS sending failed with status: %s", resp.Status)
	}

	var response struct {
		Errors bool `json:"errors"`
		Items  []struct {
			Code int `json:"code"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if response.Errors || len(response.Items) == 0 || response.Items[0].Code != 201 {
		return nil, fmt.Errorf("failed to send SMS: %v", response)
	}

	return &VerificationResponse{
		Code:      code,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *SMSService) generateCode() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := ""

	for range s.cfg.CodeLength {
		code += fmt.Sprintf("%d", r.Intn(10))
	}

	return code
}

func (s *SMSService) CheckVerificationCode(ctx context.Context, check VerificationCheck) (*VerificationResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, exists := s.verification[check.Phone]
	if !exists {
		return &VerificationResult{Success: false}, nil
	}

	if data.Code != check.Code || time.Now().After(data.ExpiresAt) {
		return &VerificationResult{Success: false}, nil
	}

	delete(s.verification, check.Phone)

	return &VerificationResult{
		Success:    true,
		VerifiedAt: time.Now(),
		Phone:      check.Phone,
	}, nil
}
