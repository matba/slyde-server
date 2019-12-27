package email

// Cache The interface that represents a generic cache
type EmailSender interface {
	// Add a key and value with a timout to cache
	SendEmail(email string, subject string, body string) error
}
