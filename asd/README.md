# CYPHONIC Authentication Service Daemon

CYPHONIC 認証サービス

## Description

CYPHONIC の認証サービス

AS では以下のことを行います

- DeviceID, Password を用いた認証
- 証明書を用いた認証
- Single Sign On(WIP)
- CYPHONIC node に NMS の IP アドレス，共通鍵，FQDN(通信用)を渡す

## Requirement

| 言語/フレームワーク | バージョン |
| :------------------ | ---------: |
| go                  |     1.23.4 |
| CockroachDB         |     23.1.8 |
| Docker              |   20.10.17 |
| docker compose      |      2.7.0 |

## Usage

1. make
2. make run

## Contribution

1. [Fork it](https://github.com/Pluslab/cyphonic/asd/fork)
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create new Pull Request
