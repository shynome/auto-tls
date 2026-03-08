# Changelog

## [0.3.0] - 2026-03-08

- 添加: `/api/ask-cert/always-ok` 接口, 满足 caddy 对泛域名的配置

## [0.2.3] - 2026-03-08

- 优化: 构建脚本

## [0.2.2] - 2026-03-08

- 修复: 阿里云的泛域名格式是 `.a.example` 而不是 `*.a.example`, 兼容它

## [0.2.1] - 2025-12-21

- 修正: 新 ACME CA 应当是 LiteSSL

## [0.2.0] - 2025-12-18

- 添加: ACME CA TrustAsia 的支持

## [0.1.1] - 2025-11-10

- 修复: 临时使用自己的 alidns 库, 等 alidns 更新修复了[错误](https://github.com/libdns/alidns/issues/9)再切回去

## [0.1.0] - 2025-11-10

- 添加: alidns 的支持

## [0.0.2] - 2025-11-07

- 修复: docker entrypoint

## [0.0.1] - 2025-11-07

成功实现了申请证书并部署到阿里云的自动化
