package smsregistration

import "time"

const cleanupPeriod = 5 * time.Minute

type Config struct {
	BaseURL       string
	AuthToken     string
	SenderName    string
	CallbackURL   string
	CodeLength    int
	CodeValidity  time.Duration
	CleanupPeriod time.Duration
}
