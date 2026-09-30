# CYPHONIC Node Daemon

CYPHONIC Node Daemon

## Description

CYPHONIC Node のデーモン

noded では以下のことを行います

- Login Request の送信
- Node Information の登録、更新
- 経路構築要求
- トンネル通信
- 仮想インターフェースの作成
- DNS パケットの処理
- IP アドレスの変更を検知

## Requirement

| 言語/フレームワーク | バージョン |
| :------------------ | ---------: |
| go                  |     1.23.3 |
| Docker              |   20.10.17 |
| docker compose      |      2.7.0 |
| just                |     1.38.0 |

## Usage

```shell
export AUTHENTICATION_SERVICE_HOST=as.local.cyphonic.org
export AUTHENTICATION_SERVICE_PORT=4501
export AUTHENTICATION_SERVICE_CA_FILE_PATH=ca.pem
export SELF_DESIRED_FQDN=test-self-desired-fqdn.cyphonic.org
go run ./cmd/main.go run
```

## Contribution

1. [Fork it](https://github.com/Pluslab/cyphonic/noded/fork)
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. reate new Pull Request
