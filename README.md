# CYPHONIC

[![License](https://img.shields.io/badge/license-MIT-orange.svg)](https://opensource.org/licenses/MIT)
[![Go](https://img.shields.io/badge/Go-1.23.4-blue.svg)](https://github.com/Pluslab/cyphonic)

[![as-reviewdog](https://github.com/Pluslab/cyphonic/actions/workflows/as-reviewdog.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/as-reviewdog.yaml)
[![nms-reviewdog](https://github.com/Pluslab/cyphonic/actions/workflows/nms-reviewdog.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/nms-reviewdog.yaml)
[![trs-reviewdog](https://github.com/Pluslab/cyphonic/actions/workflows/trs-reviewdog.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/trs-reviewdog.yaml)
[![node-reviewdog](https://github.com/Pluslab/cyphonic/actions/workflows/node-reviewdog.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/node-reviewdog.yaml)

[![GitHub Actions](https://github.com/Pluslab/cyphonic/actions/workflows/go.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/go.yaml)
[![CI](https://github.com/Pluslab/cyphonic/actions/workflows/ci.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/ci.yaml)

[![build-as](https://github.com/Pluslab/cyphonic/actions/workflows/build-as.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/build-as.yaml)
[![build-nms](https://github.com/Pluslab/cyphonic/actions/workflows/build-nms.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/build-nms.yaml)
[![build-trs](https://github.com/Pluslab/cyphonic/actions/workflows/build-trs.yaml/badge.svg)](https://github.com/Pluslab/cyphonic/actions/workflows/build-trs.yaml)

## Concepts of CYPHONIC

CYPHONIC (CYber PHysical Overlay Network over Internet Communication) is a proprietary protocol of Naito Laboratory (Pluslab).

CYPHONIC is a proven overlay network protocol supporting secure IP mobility technology that works on both IPv4 and IPv6 networks.

CYPHONIC provides secure and continuous end-to-end communication including encryption between end-nodes, NAT traversal in IPv4 networks, continuous unbroken IP mobility anywhere, and full connectivity in IPv4/IPv6 networks.

## Descripiton

The cloud service has several functions such as Authentication Service (AS), Node Management Service (NMS), and Tunnel Relay Service (TRS).

AS provides the authentification process for CYPHONIC nodes.

Additionally, it also assigns a Fully Qualified Domain Name (FQDN) as an identifier of a CYPHONIC node and distributes a shared encryption key for communication between NMS and CYPHONIC nodes.

NMS is a managing function for each CYPHONIC nodes.

It stores information about the network location of each CYPHONIC node and decides a tunnel establishing process according to the stored information.

TRS provides a relay service for communication between IPv4 and IPv6 networks and communication between private networks.

## Cloning

### Requirements

| Language/FrameWork |  Version |
| :----------------- | -------: |
| go                 |   1.23.3 |
| asdf               |   0.10.2 |
| pre-commit         |   2.19.0 |
| MariaDB            |  10.10.2 |
| Docker Engine      | 20.10.21 |
| Compose            |   2.13.0 |

### Downloading

```shell
$ git clone git@github.com:Pluslab/cyphonic.git
$ cd cyphonic
```

### Initial setting: plugins

```shell
### brew install
$ brew update && brew install asdf
```

- <https://asdf-vm.com/guide/getting-started.html>

```shell
### plugins install
$ make plugin-install
```

### Initial setting: pre-commit

```shell
### pre-commit install
$ brew update && $ brew install pre-commit
```

```shell
$ make precommit-install
```

## Documentation

- <https://cyphonic.esa.io/>

## Supported Platform

### Supported features

Nothing...

## Project managers

- [Issues](https://github.com/Pluslab/cyphonic/issues)
- [Pull Request](https://github.com/Pluslab/cyphonic/pulls)
- [Projects](https://github.com/Pluslab/cyphonic/projects/1)
- [Wiki](https://github.com/Pluslab/cyphonic/wiki)

## Contributing

Bug reports and pull requests are welcome on GitHub at [https://github.com/Pluslab/cyphonic](https://github.com/Pluslab/cyphonic).

This project is intended to be a safe, welcoming space for collaboration, and contributors are expected to adhere to the [Contributor Covenant](http://contributor-covenant.org) code of conduct.

## Code of Conduct

Everyone interacting in this project’s codebases, issue trackers, chat rooms and mailing lists is expected to follow the [code of conduct](./.github/CODE_OF_CONDUCT.md).

## Thanks to the contributors of CYPHONIC!

<a href="https://github.com/yoshikawa"><img src="https://avatars.githubusercontent.com/u/18498244?v=4" title="yoshikawa" width="80" height="80"></a>
<a href="https://github.com/GotoRen"><img src="https://avatars.githubusercontent.com/u/63791288?v=4" title="GotoRen" width="80" height="80"></a>
<a href="https://github.com/Matama091"><img src="https://avatars.githubusercontent.com/u/50354281?v=4" title="Matama091" width="80" height="80"></a>
<a href="https://github.com/komuran"><img src="https://avatars.githubusercontent.com/u/22931430?v=4" title="komuran" width="80" height="80"></a>
<a href="https://github.com/hihumikan"><img src="https://avatars.githubusercontent.com/u/26848713?v=4" title="hihumikan" width="80" height="80"></a>
<a href="https://github.com/mitsu3s"><img src="https://avatars.githubusercontent.com/u/119998577?v=4" title="mitsu3s" width="80" height="80"></a>
<a href="https://github.com/sonarAIT"><img src="https://avatars.githubusercontent.com/u/67499131?v=4" title="sonarAIT" width="80" height="80"></a>
<a href="https://github.com/Lium1126"><img src="https://avatars.githubusercontent.com/u/75057624?v=4" title="Lium1126" width="80" height="80"></a>
<a href="https://github.com/Miz32"><img src="https://avatars.githubusercontent.com/u/111434070?v=4" title="Miz32" width="80" height="80"></a>
<a href="https://github.com/Seidy-u"><img src="https://avatars.githubusercontent.com/u/85110663?v=4" title="Seidy-u" width="80" height="80"></a>
<a href="https://github.com/sudamichiyo"><img src="https://avatars.githubusercontent.com/u/101534993?v=4" title="sudamichiyo" width="80" height="80"></a>
<a href="https://github.com/Minmin067"><img src="https://avatars.githubusercontent.com/u/89015418?v=4" title="Minmin067" width="80" height="80"></a>
<a href="https://github.com/MiyamuP"><img src="https://avatars.githubusercontent.com/u/75057367?v=4" title="MiyamuP" width="80" height="80"></a>
<a href="https://github.com/hyouhyan"><img src="https://avatars.githubusercontent.com/u/76419486?v=4" title="Hyouhyan" width="80" height="80"></a>
<a href="https://github.com/mukpan"><img src="https://avatars.githubusercontent.com/u/109339477?v=4" title="Mukpan" width="80" height="80"></a>

## LICENSE

[MIT License](./LICENSE)
