# gipher
[![Build Status](https://travis-ci.org/morikuni/gipher.svg?branch=master)](https://travis-ci.org/morikuni/gipher)

gipher encrypts/decrypts structured text by password or aws-kms.

plaintext, json, yaml, and toml are supported.



## Usage

password

```
$ cat test.json
{
    "aaa": "aaa",
    "bbb": 111,
    "ccc": {
        "ddd": 222,
        "eee": "eee",
        "fff": "fff"
    }
}

$ gipher encrypt \
  --format json \
  -f test.json \
  --pattern ccc | jq
password:
{
  "aaa": "aaa",
  "bbb": 111,
  "ccc": {
    "ddd": "K0A/f1sRtp4S+N3kR6lzqYtbkEMYVSdZKeTPy1Wy",
    "eee": "l0LzhRzjhQtNaTV9K0I3AOSjD1iz9mblhas=",
    "fff": "Exbc9NPnNEI8YviY5dxP+bL6kX88ELap2NU="
  }
}

% gipher decrypt \
  --format json \
  -f encrypted.json \
  --pattern ccc | jq
password:
{
  "aaa": "aaa",
  "bbb": 111,
  "ccc": {
    "ddd": 222,
    "eee": "eee",
    "fff": "fff"
  }
}
```

aws-kms

```
$ AWS_PROFILE=default gipher encrypt \
  --format json \
  -f test.json \
  --pattern ccc \
  --cryptor aws-kms \
  --aws-region ap-northeast-1 \
  --aws-key-id alias/test | jq
{
  "aaa": "aaa",
  "bbb": 111,
  "ccc": {
    "ddd": "AQECAHgFgSrBGtkzwv+6O00BGF+UANW5TVR8ZU9AZNzY3rHwJAAAAGwwagYJKoZIhvcNAQcGoF0wWwIBADBWBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDKIkqftKQtB/HXLpGwIBEIAp4xqp5lcku4UouJ2SnKZBD773pzT8QptKY1b1PpsP1mMDhmclGqO/LN0=",
    "eee": "AQECAHgFgSrBGtkzwv+6O00BGF+UANW5TVR8ZU9AZNzY3rHwJAAAAGgwZgYJKoZIhvcNAQcGoFkwVwIBADBSBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDApjQ5SA15J08L7++AIBEIAlfKUxD8gpe5t1cHQHeYOE5SgEMPy2fU+iDnQL9e9xPBURbHYsCw==",
    "fff": "AQECAHgFgSrBGtkzwv+6O00BGF+UANW5TVR8ZU9AZNzY3rHwJAAAAGgwZgYJKoZIhvcNAQcGoFkwVwIBADBSBgkqhkiG9w0BBwEwHgYJYIZIAWUDBAEuMBEEDPBRIWYH3xZ4a3CRxQIBEIAli7hPcTXkkxF+lJrMhKD4DekZyiiz4vbxz6zfG0dPCPaXp+xOdQ=="
  }
}

$ AWS_PROFILE=default gipher decrypt \
  --format json \
  -f encrypted.json \
  --pattern ccc \
  --cryptor aws-kms \
  --aws-region ap-northeast-1 | jq
{
  "aaa": "aaa",
  "bbb": 111,
  "ccc": {
    "ddd": 222,
    "eee": "eee",
    "fff": "fff"
  }
}
```
