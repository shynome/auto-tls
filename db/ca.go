package db

import "github.com/caddyserver/certmagic"

const (
	CALetsEncryptProduction = "LetsEncrypt"
	CALetsEncryptStaging    = "LetsEncryptStaging"
	CAZeroSSLProduction     = "ZeroSSL"
	CALiteSSL               = "LiteSSL"
	CAGoogleTrustProduction = "GoogleTrust"
	CAGoogleTrustStaging    = "GoogleTrustStaging"

	// Deprecated: 错误的使用 TrustAsia 作为品牌名, 更正为 LiteSSL
	CATrustAsia = "TrustAsia"
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
	case CALiteSSL, CATrustAsia:
		return "https://acme.litessl.com/acme/v2/directory"
	}
	return ""
}
