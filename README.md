# 证书申请与部署工具

# 如何使用

等我录个视频教程吧

# 部署目标

## Caddy

通过 [tls.get_certificate.http](https://caddyserver.com/docs/caddyfile/directives/tls#http-1) 获取证书

```txt
local.tls-share-test.shynome.com {
  tls {
    get_certificate http http://auto-tls.lo.shynome.com:8090/api/cert/{domain_id}?token=6
  }
  respond "local.tls-share-test.shynome.com"
}
```

## 阿里云

要申请 AccessKey, 配置文件结构是:

```json
{ "key": "AccessKey ID", "secret": "AccessKey Secret" }
```

需要的权限有

| 权限                         | 原因            |
| ---------------------------- | --------------- |
| `AliyunYundunCertFullAccess` | 上传证书(必需)  |
| `AliyunDCDNFullAccess`       | 部署证书到 DCDN |
| `AliyunCDNFullAccess`        | 部署证书到 CDN  |

# Todo

一些缺失的功能(我暂时用不上), 其实也是付费需求, 如果你公司(可开票)或个人有需求可向我支付对应的价格以便我进行开发, 邮箱: [shynome@remoon.cn](mailto:shynome@remoon.cn)

- [ ] 更多 dns_providers 的支持 (300 元)
  - [x] cloudflare
  - [ ] 阿里云
- [ ] 更多部署目标支持
  - [ ] 阿里云
    - [x] CDN
    - [x] DCDN
    - [ ] ESA (300 元)
  - [ ] 腾讯云 (2_000 元)
