package client

// CA represents a certificate authority
type CA struct {
	CommonName string
	Cert       string // PEM-encoded
}
