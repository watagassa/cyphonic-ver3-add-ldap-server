\c cyphonic cyphonic;

INSERT INTO
  accounts(id, email, password, permission)
VALUES
  (
    1,
    'test01@cyphonic.org',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04',
    1
  ),
  (
    2,
    'test02@cyphonic.org',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04',
    1
  ),
  (
    3,
    'test03@cyphonic.org',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04',
    1
  ),
  (
    4,
    'test04@cyphonic.org',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04',
    1
  );

INSERT INTO
  devices(
    device_id,
    account_id,
    device_name,
    fqdn,
    device_type_id,
    group_profile,
    adapter_flag,
    general_node_flag,
    inter_node_certification_flag,
    status
  )
VALUES
  (
    'f6efb4e0355cae5a37ddd58a286d4bcbe6ff4ca585c67cb02f263aa3',
    1,
    'node',
    'node01.cyphonic.org',
    1,
    1,
    false,
    false,
    false,
	  1
  ),
  (
    'f7b8c180a4d573cb55479f76377f35fbbe58453362736f6640c2612a',
    2,
    'node2',
    'node02.cyphonic.org',
    1,
    1,
    false,
    false,
    false,
  	1
  ),
  (
    '204df901511616cf3b9eba8df20003897e54bb3fceaaa3bf4ea85d50',
    3,
    'node3',
    'node03.cyphonic.org',
    1,
    2,
    false,
    false,
    false,
  	1
  ),
  (
    '0cd80324265f346233dfcf4318725b141be4d59676fb2836e684c560',
    4,
    'node4',
    'node04.cyphonic.org',
    1,
    2,
    false,
    false,
    false,
  	1
  );

INSERT INTO
  device_passwords(device_id, password)
VALUES
  (
    'f6efb4e0355cae5a37ddd58a286d4bcbe6ff4ca585c67cb02f263aa3',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04'
  ),
  (
    'f7b8c180a4d573cb55479f76377f35fbbe58453362736f6640c2612a',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04'
  ),
  (
    '204df901511616cf3b9eba8df20003897e54bb3fceaaa3bf4ea85d50',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04'
  ),
  (
    '0cd80324265f346233dfcf4318725b141be4d59676fb2836e684c560',
    '99425113951d8c9e0138571f8ca74ec42d0e91a501a57887ff39c476b57463abde67b26eb90426d06c197193677b8da5f123ac88fce3c75f03915818d0732a04'
  );

INSERT INTO
  virtual_ip_addresses(device_id, virtual_ipv4, virtual_ipv4_netmask, virtual_ipv6, virtual_ipv6_prefix)
VALUES
  (
    'f6efb4e0355cae5a37ddd58a286d4bcbe6ff4ca585c67cb02f263aa3',
    decode('c6120065', 'hex'),
    decode('ffff0000', 'hex'),
    decode('20010db8c0ffee000000000000000065', 'hex'),
    decode('ffffffffffffffff0000000000000000', 'hex')
  ),
  (
    'f7b8c180a4d573cb55479f76377f35fbbe58453362736f6640c2612a',
    decode('c6120066', 'hex'),
    decode('ffff0000', 'hex'),
    decode('20010db8c0ffee000000000000000066', 'hex'),
    decode('ffffffffffffffff0000000000000000', 'hex')
  ),
  (
    '204df901511616cf3b9eba8df20003897e54bb3fceaaa3bf4ea85d50',
    decode('c6120067', 'hex'),
    decode('ffff0000', 'hex'),
    decode('20010db8c0ffee000000000000000067', 'hex'),
    decode('ffffffffffffffff0000000000000000', 'hex')
  ),
  (
    '0cd80324265f346233dfcf4318725b141be4d59676fb2836e684c560',
    decode('c6120068', 'hex'),
    decode('ffff0000', 'hex'),
    decode('20010db8c0ffee000000000000000068', 'hex'),
    decode('ffffffffffffffff0000000000000000', 'hex')
  );

INSERT INTO
  fqdn_aliases(device_id, fqdn_alias)
VALUES
  (
    'f6efb4e0355cae5a37ddd58a286d4bcbe6ff4ca585c67cb02f263aa3',
    'alias01.cyphonic.org'
  ),
  (
    'f7b8c180a4d573cb55479f76377f35fbbe58453362736f6640c2612a',
    'alias02.cyphonic.org'
  ),
  (
    '204df901511616cf3b9eba8df20003897e54bb3fceaaa3bf4ea85d50',
    'alias03.cyphonic.org'
  ),
  (
    '0cd80324265f346233dfcf4318725b141be4d59676fb2836e684c560',
    'alias04.cyphonic.org'
  );
