package app

import (
	"fmt"
)

type AppConfig struct {
	EnvNordigenAPIKey       string
	EnvNordigenRefreshToken string
	EnvNordigenAccountID    string
	EnvNordigenSecretId     string
	EnvNordigenSecretKey    string
	EnvPretixAPIKey         string
	EnvPretixEventSlug      string
	EnvPretixOrganizerSlug  string
	EnvPretixBaseUrl        string
	SmtpPort                string
	SmtpServer              string
	SmtpUsername            string
	SmtpPassword            string
	SenderEmail             string
	RecipientEmail          string
}

func (c *AppConfig) init() {
	// Extra no error check here. Default should be defined environment variables. This is only for development
	godotenv.Load(".Env")

	c.EnvNordigenAPIKey = getenv("NORDIGEN_API_KEY")
	c.EnvNordigenRefreshToken = getenv("NORDIGEN_REFRESH_KEY")
	c.EnvNordigenSecretId = getenv("NORDIGEN_SECRET_ID")
	c.EnvNordigenSecretKey = getenv("NORDIGEN_SECRET_KEY")
	c.EnvNordigenAccountID = getenv("NORDIGEN_ACCOUNT_ID")

	c.EnvPretixAPIKey = getenv("PRETIX_API_KEY")
	c.EnvPretixEventSlug = getenv("PRETIX_EVENT_SLUG")
	c.EnvPretixOrganizerSlug = getenv("PRETIX_ORGANIZER_SLUG")
	c.EnvPretixBaseUrl = getenv("PRETIX_BASE_URL")

	c.SmtpPort = getenv("SMTP_PORT")
	c.SmtpServer = getenv("SMTP_SERVER")
	c.SmtpUsername = getenv("SMTP_USER")
	c.SmtpPassword = getenv("SMTP_PASSWORD")
	c.SenderEmail = getenv("SENDER_MAIL")
	c.RecipientEmail = getenv("RECIPIENT_MAIL")
}

func Run() {
	initConfig()

}

func initConfig() {

}
