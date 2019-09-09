kill-mysql-connection
=====================

[![Actions Status](https://github.com/li-go/kill-mysql-connection/workflows/Go/badge.svg?branch=develop)](https://github.com/li-go/kill-mysql-connection/actions)

### How To Install

```
$ go get -u github.com/li-go/kill-mysql-connection
```

### How To Use

```
$ kill-mysql-connection -h
Usage of kill-mysql-connection:
  -config string
        config file in toml format
  -max-time int
        kill process lives longer than max-time seconds (default 100)
```

### Config (in toml format)

```
[mysql]
  host = ""
  port = 3306
  username = ""
  password = ""

[ssh_tunnel]
  use_tunnel = false
  host = ""
  port = 22
  username = ""
  password = ""

  private_key = ""
  key_passphrase = ""
```
