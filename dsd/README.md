# CYPHONIC Direction Selection Daemon

CYPHONIC Direction Selection

## Description

CYPHONIC Direction Selection

dsd では以下のことを行います

- Node のインターネットの情報(IP, NAT)などを NS に通知
- TRS の中継が不要な場合の経路構築を担います

## Requirement

| 言語/フレームワーク | バージョン |
| :------------------ | ---------: |
| go                  |     1.23.3 |
| CockroachDB         |     23.1.8 |
| Docker              |   20.10.17 |
| docker compose      |      2.7.0 |

## Usage

1. make
2. make run

## Contribution

1. [Fork it](https://github.com/Pluslab/cyphonic/fork)
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create new Pull Request
