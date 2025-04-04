package infrastructure

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type TwilioSMSService struct {
	accountSid string
	authToken  string
	fromNumber string
}

func NewTwilioSMSService() *TwilioSMSService {
	return &TwilioSMSService{
		accountSid: os.Getenv("TWILIO_ACCOUNT_SID"),
		authToken:  os.Getenv("TWILIO_AUTH_TOKEN"),
		fromNumber: os.Getenv("TWILIO_FROM_NUMBER"),
	}
}

func (s *TwilioSMSService) SendAlert(phoneNumber string, message string) error {
	endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", s.accountSid)

	data := url.Values{}
	data.Set("To", "+52"+phoneNumber)
	data.Set("From", s.fromNumber)
	data.Set("Body", message)

	req, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	req.SetBasicAuth(s.accountSid, s.authToken)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error sending SMS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("error response from Twilio: %s", resp.Status)
	}

	log.Printf("✅ Alert SMS sent to %s", phoneNumber)
	return nil
}
