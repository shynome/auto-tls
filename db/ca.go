package db

import "github.com/caddyserver/certmagic"

const (
	CALetsEncryptProduction = "LetsEncrypt"
	CALetsEncryptStaging    = "LetsEncryptStaging"
	CAZeroSSLProduction     = "ZeroSSL"
	CATrustAsia             = "TrustAsia"
	CAGoogleTrustProduction = "GoogleTrust"
	CAGoogleTrustStaging    = "GoogleTrustStaging"
)

type CA string

func MagicCA(c string) string {
	switch c {
	case CALetsEncryptProduction:
		return certmagic.LetsEncryptProductionCA
	case CALetsEncryptStaging:
		return certmagic.LetsEncryptStagingCA
	case CAZeroSSLProduction:
		return certmagic.ZeroSSLProductionCA
	case CAGoogleTrustProduction:
		return certmagic.GoogleTrustProductionCA
	case CAGoogleTrustStaging:
		return certmagic.GoogleTrustStagingCA
	case CATrustAsia:
		return "https://acme.litessl.com/acme/v2/directory"
	}
	return ""
}
