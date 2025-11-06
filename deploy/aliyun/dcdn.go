package aliyun

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/dcdn"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

type DCDN struct {
	Config
}

func NewDCDN(config Config) *DCDN {
	return &DCDN{
		Config: config,
	}
}

func (c *DCDN) getClientTry() *dcdn.Client {
	client := try.To1(dcdn.NewClientWithAccessKey(c.Region, c.Key, c.Secret))
	return client
}

func (c *DCDN) List(suffix string) (_ []dcdn.PageData, err error) {
	defer err0.Then(&err, nil, nil)

	client := c.getClientTry()

	domains, err := loop(PageSize, func(i int) ([]dcdn.PageData, error) {
		req := dcdn.CreateDescribeDcdnUserDomainsRequest()
		req.PageSize = requests.NewInteger(PageSize)
		req.PageNumber = requests.NewInteger(i)
		req.DomainStatus = "online"
		if suffix != "" {
			req.DomainSearchType = "suf_match"
			req.DomainName = suffix
		}
		resp, err := client.DescribeDcdnUserDomains(req)
		if err != nil {
			return nil, err
		}
		return resp.Domains.PageData, nil
	})
	try.To(err)

	return domains, nil
}

func (c *DCDN) Deploy(domain string, cert int) (_ *dcdn.SetDcdnDomainSSLCertificateResponse, err error) {
	defer err0.Then(&err, nil, nil)

	client := c.getClientTry()

	req := dcdn.CreateSetDcdnDomainSSLCertificateRequest()
	req.DomainName = domain
	req.CertName = domain
	req.CertId = requests.NewInteger(cert)
	req.CertType = "cas"
	req.SSLProtocol = "on"
	resp := try.To1(client.SetDcdnDomainSSLCertificate(req))

	return resp, nil
}
