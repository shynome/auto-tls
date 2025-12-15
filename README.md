# 证书申请与部署工具

项目前身 [shynome/tls-deploy](https://github.com/shynome/tls-deploy), 后面又因为证书时长要缩短到 1 年以下,
买证书也没有什么意义了, 还是上自动化工具吧, 于是做了这个集中管理证书并部署的工具

# 如何使用

```sh
# 创建数据存储目录
mkdir auto-tls
# 运行服务
docker run -d --restart always --name auto-tls -p 9443:9443 -v $PWD/auto-tls/:/app/pb_data/ shynome/auto-tls:v0.0.2
# 创建管理员帐号
docker exec auto-tls ./auto-tls superuser create admin@tls.local 12345678
```

打开浏览器管理页面: <http://127.0.0.1:9443/_/> , 接着使用上方创建的管理员帐号登录

接下来就参考下面一分钟的视频展示了(视频中出现的认证凭据均已删除)

https://github.com/user-attachments/assets/39f64d61-317c-4388-8414-9b7ec849505b

# 部署目标

## Caddy

通过 [tls.get_certificate.http](https://caddyserver.com/docs/caddyfile/directives/tls#http-1) 获取证书, [Caddy v2.5.0](https://github.com/caddyserver/caddy/pull/4541) 后支持

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


- [ ] 部署目标的字段加上部署平台的前缀
- [ ] 更多 dns_providers 的支持 (300 元)
  - [x] cloudflare
  - [x] 阿里云(alidns)
- [ ] 更多部署目标支持
  - [ ] 阿里云
    - [x] CDN
    - [x] DCDN
    - [ ] ESA (300 元)
  - [ ] 腾讯云 (2_000 元)
- [ ] 部署任务失败时发送邮件通知
