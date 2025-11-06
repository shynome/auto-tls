package tools

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"

	"github.com/caddyserver/certmagic"
)

func ToPEM(cert tls.Certificate) (_ []byte, err error) {
	body := &bytes.Buffer{}
	for _, der := range cert.Certificate {
		block := &pem.Block{Type: "CERTIFICATE", Bytes: der}
		if err := pem.Encode(body, block); err != nil {
			return nil, err
		}
	}
	key, err := certmagic.PEMEncodePrivateKey(cert.PrivateKey)
	if err != nil {
		return nil, err
	}
	body.Write(key)
	return body.Bytes(), nil
}
