package smsregistration

import "time"

type Config struct {
	BaseURL              string
	AuthToken            string
	SenderName           string
	CodeLength           int
	CodeValidityDuration time.Duration
	CleanupPeriod        time.Duration
}
