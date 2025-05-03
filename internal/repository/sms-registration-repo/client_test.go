package smsregistrationrepo_test

import (
	"testing"

	smsregistrationrepo "github.com/romanpitatelev/clothing-service/internal/repository/sms-registration-repo"
	"github.com/stretchr/testify/require"
)

func TestSMSService_SendOTP(t *testing.T) {
	client := smsregistrationrepo.New(smsregistrationrepo.Config{
		Host:     "gate.smsaero.ru",
		Schema:   "https",
		Email:    "rpitatelev@gmail.com",
		ApiKey:   "o7KDkzhEcTFceryZLZ2xZcs3muTWgi-P",
		Sender:   "SMS Aero",
		TestMode: true,
	})

	err := client.SendOTP(t.Context(), "79031355530", "biba")
	require.NoError(t, err)
}
