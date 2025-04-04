package output

type SMSService interface {
	SendAlert(phoneNumber string, message string) error
}
