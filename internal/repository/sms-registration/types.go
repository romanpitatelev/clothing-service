package smsregistration

import "time"

type VerificationResponse struct {
	Code      string    `json:"code"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type VerificationCheck struct {
	Phone string `json:"phoneNumber"`
	Code  string `json:"code"`
}

type VerificationResult struct {
	Success    bool      `json:"success"`
	VerifiedAt time.Time `json:"verifiedAt"`
	Phone      string    `json:"phone"`
}
