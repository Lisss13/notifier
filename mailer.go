package notifier

import (
	"time"

	sendgrid "github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

	"github.com/powerman/structlog"
)

const (
	errMsgBadRecipientCode = 501

	minRetryDelayDefault = 15 * time.Second
	maxRetryDelayDefault = 15 * time.Minute
)

// Mailer - struct with config and methods for sending emails.
// Wrapping the sendgrid api
type Mailer struct {
	log            *structlog.Logger
	config         ConfigMailer
	sendGridClient *sendgrid.Client
}

// ConfigMailer - config structure for sending email.
type ConfigMailer struct {
	From     string
	NameFrom string
	Pass     string // Pass is sendgrid api key.

	MinRetryDelay time.Duration
	MaxRetryDelay time.Duration
}

// Mail contains mail struct.
type Mail struct {
	message *mail.SGMailV3
}

// NewMailer - create new instance Mailer with checking the delays for correctness.
func NewMailer(log *structlog.Logger, config ConfigMailer) *Mailer {
	if config.MinRetryDelay > config.MaxRetryDelay {
		config.MaxRetryDelay = config.MinRetryDelay
	}
	if config.MinRetryDelay <= 0 {
		config.MinRetryDelay = minRetryDelayDefault
	}
	if config.MaxRetryDelay <= 0 {
		config.MaxRetryDelay = maxRetryDelayDefault
	}

	return &Mailer{
		log:            log,
		config:         config,
		sendGridClient: sendgrid.NewSendClient(config.Pass),
	}
}

// SendEmail sends mail to email and tries to send it until the email is sent or an error occurs.
func (m *Mailer) SendEmail(email string, mail Mail) {
	for delay := NewExpDelay(m.config.MinRetryDelay, m.config.MaxRetryDelay); ; delay.Sleep() {
		response, err := m.sendGridClient.Send(mail.message)
		if err == nil && response.StatusCode == 202 {
			m.log.Info("send email", "to", email, "from", m.config.From, "status code", response.StatusCode, "response body", response.Body)
			break
		}
		if err != nil {
			m.log.PrintErr("failed to send email", "to", email, "from", m.config.From, "err", err)
		} else {
			m.log.PrintErr("failed to send email", "to", email, "from", m.config.From, "status code", response.StatusCode, "response body", response.Body)
			if response.StatusCode == errMsgBadRecipientCode {
				break
			}
		}
	}
}

// CreateEmail gets email props and returns prepared mail for SendEmail method.
func (m *Mailer) CreateEmail(to, subject, bodyMessage string) Mail {
	fromEmail := mail.NewEmail(m.config.NameFrom, m.config.From)
	toEmail := mail.NewEmail(to, to)

	return Mail{
		message: mail.NewSingleEmail(fromEmail, subject, toEmail, "", bodyMessage),
	}
}
