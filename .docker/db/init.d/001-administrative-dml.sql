\c cyphonic cyphonic;

INSERT INTO
  device_types (id, device_type)
VALUES
  (1, 'linux'),
  (2, 'windows'),
  (3, 'darwin'),
  (4, 'ios'),
  (5, 'android');

INSERT INTO
  group_profiles(
    bit_shift,
    group_profile_name
  )
VALUES
  (0, 'AIT/Pluslab'),
  (1, 'Meijo-Univ/UCLab');
