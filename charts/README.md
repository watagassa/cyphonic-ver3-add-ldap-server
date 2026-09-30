# charts

Define resource for CYPHONIC cloud microservices in Helm chart.

## Version

| Tools   | Version |
| :------ | ------: |
| kubectl | v1.25.3 |
| helm    |  v3.8.1 |

## CYPHONIC Cloud: Micro service

| Micro service |   Chart name | apiVersion |                                    Overview |              Components |
| :------------ | -----------: | ---------: | ------------------------------------------: | ----------------------: |
| AS            |  cyphonic-as |         v2 |             Authenticate all CYPHONIC nodes |  Authentication Service |
| NMS           | cyphonic-nms |         v2 | Manage communication for all CYPHONIC nodes | Node Management Service |
| TRS           | cyphonic-trs |         v2 |        Relay processing of CYPHONIC packets |    Tunnel Relay Service |
| API           | cyphonic-api |         v2 |      CYPHONIC Cloud software controller API |      Controller backend |
| WEB           | cyphonic-web |         v2 |                           CYPHONIC Web page |               Web front |
| OPS           | cyphonic-ops |         v2 |               CYPHONIC Cloud Infrastructure |      　 RootCA, SSO-IdP |
