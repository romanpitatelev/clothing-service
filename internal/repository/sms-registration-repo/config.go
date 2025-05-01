package smsregistrationrepo

import "time"

type Config struct {
	Email                string
	ApiKey               string
	Sender               string
	CodeLength           int
	CodeValidityDuration time.Duration
}
