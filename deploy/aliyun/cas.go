package aliyun

import (
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/aliyun/alibaba-cloud-sdk-go/services/cas"
	"github.com/shynome/err0"
	"github.com/shynome/err0/try"
)

type CAS struct {
	Config
}

func NewCas(config Config) *CAS {
	return &CAS{
		Config: config,
	}
}

func (c *CAS) getClientTry() *cas.Client {
	client := try.To1(cas.NewClientWithAccessKey(c.Region, c.Key, c.Secret))
	return client
}

func (c *CAS) Upload(pem []byte, name string) (_ *cas.UploadUserCertificateResponse, err error) {
	defer err0.Then(&err, nil, nil)

	client := c.getClientTry()
	req := cas.CreateUploadUserCertificateRequest()
	req.Name = name
	req.Cert = string(pem)
	req.Key = string(pem)
	resp := try.To1(client.UploadUserCertificate(req))

	return resp, nil
}

func (c *CAS) First(name string) (_ *cas.CertificateOrderListItem, err error) {
	defer err0.Then(&err, nil, nil)

	client := c.getClientTry()
	req := cas.CreateListUserCertificateOrderRequest()
	req.Keyword = name
	req.OrderType = "UPLOAD"
	req.ShowSize = requests.NewInteger(1)
	resp := try.To1(client.ListUserCertificateOrder(req))
	l := resp.CertificateOrderList
	if len(l) < 1 {
		return nil, nil
	}
	item := l[0]
	if item.Name != name {
		return nil, nil
	}
	return &item, nil
}
