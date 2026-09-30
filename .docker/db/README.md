# Database schema

## Rules

- Naming convention
  - When defining DDL
    `<prefix>-<any name>-<ddl>.sql`
  - When defining DML
    - `<prefix>-<any name>-<dml>.sql`
- About prefixes
  - The prefix manages dependencies when adding schemas.
  - For example, when A depends on B, the file names will be `001-B-dml.sql` and `002-A-dml.sql` respectively.

## Sample DML

### `002-sample-info-dml.sql`

### `003-adapter-sample-dml.sql`

```
DATUM OF "accounts"
- adapter control account: test10@cyphonic.org (cyphonic)
```

```
DATUM OF "devices"
- cyphonic-adapter-01: 2bac8cfcfa0a0e8fe7655cff0fec896866eff52e94efce030be15337 (test10@cyphonic.org)
- general-node-01: 61df7869c6ea4fca675a6272d25e80cb841d0c4ce1b2b2b7e0815bb2 (test11@cyphonic.org)
- general-node-02: 395c73de6723d246159f3c41160363da996038a34145a4ee8752a0aa (test12@cyphonic.org)
- general-node-03: 3ef68dac78ec4bce91d2cc31219e4f7b3911148453d1f14f7fc9f701 (test13@cyphonic.org)
- general-node-04: 4f0b1fb5663eb66f0a5155606f91e37f6333263716df86f451a273c0 (test14@cyphonic.org)
- general-node-05: ca4698e4ba6b2ff74fcab04584221d8a1534613707493ab51d0084b2 (test15@cyphonic.org)
```

```
DATUM OF "virtual_ip_addresses"
- cyphonic-adapter-01: 198.18.0.1/2001:db8:c0ff:ee00::1
- general-node-01: 198.18.32.5 / 2001:db8:c0ff:ee00:118:910:a215:905
- general-node-02: 198.18.32.6 / 2001:db8:c0ff:ee00:118:910:a215:906
- general-node-03: 198.18.32.7 / 2001:db8:c0ff:ee00:118:910:a215:907
- general-node-04: 198.18.32.8 / 2001:db8:c0ff:ee00:118:910:a215:908
- general-node-05: 198.18.32.9 / 2001:db8:c0ff:ee00:118:910:a215:909
```

```
DATUM OF "general_devices"
- general-node-01: dc:a6:32:b2:83:10 (61df7869c6ea4fca675a6272d25e80cb841d0c4ce1b2b2b7e0815bb2)
- general-node-02: dc:a6:32:bf:8b:a7 (395c73de6723d246159f3c41160363da996038a34145a4ee8752a0aa)
- general-node-03: dc:a6:32:bf:8b:5c (3ef68dac78ec4bce91d2cc31219e4f7b3911148453d1f14f7fc9f701)
- general-node-04: dc:a6:32:bf:8b:aa (4f0b1fb5663eb66f0a5155606f91e37f6333263716df86f451a273c0)
- general-node-05: dc:a6:32:bf:8b:77 (ca4698e4ba6b2ff74fcab04584221d8a1534613707493ab51d0084b2)
```
