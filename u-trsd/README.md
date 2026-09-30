# CYPHONIC Tunnel Relay Service Daemon

CYPHONIC トンネルリレーサービス

## Description

CYPHONIC のトンネルリレーサービス

TRS では以下のことを行います

- CYPHONIC ノード同士が、相互通信不可能な場合に、TRS を介して通信を行います

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

1. [Fork it](https://github.com/Pluslab/cyphonic/u-trsd/fork)
2. Create your feature branch
3. Commit your changes
4. Push to the branch
5. Create new Pull Request
