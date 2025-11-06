package aliyun

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cdn"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

type CDN struct {
	Config
}

func NewCDN(config Config) *CDN {
	return &CDN{
		Config: config,
	}
}

func (c *CDN) getClientTry() *cdn.Client {
	client := try.To1(cdn.NewClientWithAccessKey(c.Region, c.Key, c.Secret))
	return client
}

func (c *CDN) List(suffix string) (_ []cdn.PageData, err error) {
	defer err0.Then(&err, nil, nil)

	client := c.getClientTry()

	domains, err := loop(PageSize, func(i int) ([]cdn.PageData, error) {
		req := cdn.CreateDescribeUserDomainsRequest()
		req.PageSize = requests.NewInteger(PageSize)
		req.PageNumber = requests.NewInteger(i)
		req.DomainStatus = "online"
		if suffix != "" {
			req.DomainSearchType = "suf_match"
			req.DomainName = suffix
		}
		resp, err := client.DescribeUserDomains(req)
		if err != nil {
			return nil, err
		}
		return resp.Domains.PageData, nil
	})
	try.To(err)

	return domains, nil
}

func (c *CDN) Deploy(domain string, cert int) (_ *cdn.SetCdnDomainSSLCertificateResponse, err error) {
	defer err0.Then(&err, nil, nil)

	client := c.getClientTry()

	req := cdn.CreateSetCdnDomainSSLCertificateRequest()
	req.DomainName = domain
	req.CertName = domain
	req.CertId = requests.NewInteger(cert)
	req.CertType = "cas"
	req.SSLProtocol = "on"
	resp := try.To1(client.SetCdnDomainSSLCertificate(req))

	return resp, nil
}
