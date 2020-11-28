package util

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

type TLS struct {
	CertificateFile          string
	PrivateKeyFile           string
	CertificateAuthorityFile string
	SkipVerifyCertificate    bool
}

func (_this *TLS) Validate() error {
	if _this.CertificateFile == "" || _this.PrivateKeyFile == "" {
		return fmt.Errorf("client cert or client key must not be empty")
	}

	cf, err := GetAbsolutelyPath(_this.CertificateFile)
	if err != nil {
		return err
	}

	pf, err := GetAbsolutelyPath(_this.PrivateKeyFile)
	if err != nil {
		return err
	}

	if _this.CertificateAuthorityFile != "" {
		caf, err := GetAbsolutelyPath(_this.CertificateAuthorityFile)
		if err != nil {
			return err
		}
		_this.CertificateAuthorityFile = caf
	}

	_this.CertificateFile = cf
	_this.PrivateKeyFile = pf
	return nil
}

// NewTLS
func NewTLS(c *TLS) (*tls.Config, error) {
	err := c.Validate()
	if err != nil {
		return nil, err
	}

	cert, err := tls.LoadX509KeyPair(c.CertificateFile, c.PrivateKeyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
	}

	// Load CA certificate
	if c.CertificateAuthorityFile != "" {
		caCert, err := ioutil.ReadFile(c.CertificateAuthorityFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	tlsConfig.InsecureSkipVerify = c.SkipVerifyCertificate

	return tlsConfig, nil
}
