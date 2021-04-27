# About

**aws-route53-ddns-client** is an AWS Route 53 DDNS client written in PHP & Go. If you can run PHP or Go, need a DDNS client,
 and are willing to use AWS Route 53, then this may work for you.

## Installation

The recommended installation method is via `git`:

```shell
git clone https://github.com/rfpludwick/aws-route53-ddns-client
```

You'll want to copy the following files in the `config/` directory:

```shell
cd config
cp aws_credentials.dist.ini aws_credentials.ini
cp config.dist.ini config.ini
```

You'll want to fill in the copied INI files with your own values, which will be described in the **Setup** link below.

## Setup

### PHP Setup

This repository borrows greatly from the implementation at
[Flynsarmy](https://www.flynsarmy.com/2015/12/setting-up-dynamic-dns-to-your-home-with-route-53/). Instead of the PHP
script detailed on that page, you can use this repository.

### Go Setup

If you want to run the Go version of the client, make sure you have Go installed and then:

```shell
go get -u github.com/aws/aws-sdk-go/...
go get gopkg.in/ini.v1
go build -o cli src/go/*
```

## Usage

### PHP Usage

Invoke in any number of ways! All assuming you are in the root project directory:

```shell
../path/to/composer execute
```

```shell
./cli.php
```

```shell
php cli.php
```

The Flynsarmy article describes running this via an HTTP call. This repository can be executed via the commandline, and
it is recommended to do that for security's sake.

### Go Usage

If you've built the executable, then just execute it:

```shell
./cli
```

## Recommendations

You should consider scheduling a job to run this on a regular basis. I'm personally using a Linux server with a
`crontab` job running once per minute.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.
