package auth

import (
	"net/http"
	"time"
)

// ClientCert holds client certificate info.
type ClientCert struct {
	Subject  string
	Issuer   string
	Serial   string
	NotAfter time.Time
}

// ClientIdentity is an mTLS auth result.
type ClientIdentity struct {
	Cert *ClientCert
}

type mtlsAuth struct {
	allowedCNs map[string]bool
}

func newMTLSAuth(cfg any) (Authenticator, error) {
	m, ok := cfg.(map[string]bool)
	if !ok || m == nil {
		m = make(map[string]bool)
	}
	return &mtlsAuth{allowedCNs: m}, nil
}

func (a *mtlsAuth) Authenticate(r *http.Request) (any, error) {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return nil, ErrUnauthorized
	}
	cert := r.TLS.PeerCertificates[0]
	cn := cert.Subject.CommonName
	if len(a.allowedCNs) > 0 && !a.allowedCNs[cn] {
		return nil, ErrUnauthorized
	}
	return &Identity{
		Subject: cn,
		Source:  "mtls",
		Claims: map[string]any{
			"issuer":    cert.Issuer.CommonName,
			"serial":    cert.SerialNumber.String(),
			"not_after": cert.NotAfter,
		},
	}, nil
}

// ClientCertFromRequest extracts the certificate (helper).
func ClientCertFromRequest(r *http.Request) *ClientCert {
	if r.TLS == nil || len(r.TLS.PeerCertificates) == 0 {
		return nil
	}
	c := r.TLS.PeerCertificates[0]
	return &ClientCert{
		Subject:  c.Subject.CommonName,
		Issuer:   c.Issuer.CommonName,
		Serial:   c.SerialNumber.String(),
		NotAfter: c.NotAfter,
	}
}
