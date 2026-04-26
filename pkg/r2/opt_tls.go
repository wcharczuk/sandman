package r2

import (
	"crypto/tls"
	"crypto/x509"
	"time"
)

// OptTLSClientConfig sets the tls config for the request.
// It will create a client, and a transport if unset.
func OptTLSClientConfig(cfg *tls.Config) Option {
	return func(r *Request) error {
		transport, err := EnsureHTTPTransport(r)
		if err != nil {
			return err
		}
		transport.TLSClientConfig = cfg
		return nil
	}
}

// OptTLSInsecureSkipVerify sets if we should skip verification.
func OptTLSInsecureSkipVerify(insecureSkipVerify bool) Option {
	return func(r *Request) error {
		transport, err := EnsureHTTPTransport(r)
		if err != nil {
			return err
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.InsecureSkipVerify = insecureSkipVerify
		return nil
	}
}

// OptTLSHandshakeTimeout sets the client transport TLSHandshakeTimeout.
func OptTLSHandshakeTimeout(d time.Duration) Option {
	return func(r *Request) error {
		transport, err := EnsureHTTPTransport(r)
		if err != nil {
			return err
		}
		transport.TLSHandshakeTimeout = d
		return nil
	}
}

// OptTLSRootCAs sets the client tls root ca pool.
func OptTLSRootCAs(pool *x509.CertPool) Option {
	return func(r *Request) error {
		transport, err := EnsureHTTPTransport(r)
		if err != nil {
			return err
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.RootCAs = pool
		return nil
	}
}

// OptTLSClientCert adds a client certificate to the request.
func OptTLSClientCert(cert tls.Certificate) Option {
	return func(r *Request) error {
		transport, err := EnsureHTTPTransport(r)
		if err != nil {
			return err
		}
		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}
		transport.TLSClientConfig.Certificates = append(transport.TLSClientConfig.Certificates, cert)
		return nil
	}
}
