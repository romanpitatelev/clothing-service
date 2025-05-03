package tests

import (
	"context"
	"encoding/json"
	"net/http"

	smsregistrationrepo "github.com/romanpitatelev/clothing-service/internal/repository/sms-registration-repo"
)

type otpResp struct {
	phone  string
	otp    string
	sender string
}

func (s *IntegrationTestSuite) runServer(ctx context.Context, bindAddr string) {
	//nolint:gosec
	server := &http.Server{
		Addr: bindAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			data, err := json.Marshal(smsregistrationrepo.Response{Success: true})
			s.Require().NoError(err)

			_, err = w.Write(data)
			s.Require().NoError(err)

			go func() {
				s.smsChan <- otpResp{
					phone:  r.URL.Query().Get("number"),
					otp:    smsregistrationrepo.ExtractOTP(r.URL.Query().Get("text")),
					sender: r.URL.Query().Get("sign"),
				}
			}()
		}),
	}
	go func() {
		<-ctx.Done()

		_ = server.Close()
	}()

	err := server.ListenAndServe()
	s.Require().ErrorIs(err, http.ErrServerClosed)
}
