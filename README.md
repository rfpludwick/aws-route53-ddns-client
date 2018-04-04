# About

**aws-route53-ddns-php-client** is an AWS Route 53 DDNS client written in PHP. If you can run PHP, need a DDNS client,
 and are willing to use AWS Route 53, then this may work for you.

# Installation

The recommended installation method is via Composer:

```bash
cd /path/to/where/this/will/live
../path/to/composer install rfpludwick/aws-route53-ddns-php-client
```

You'll want to copy the following files in the `config/` directory:

```bash
cd config
cp aws_credentials.dist.ini aws_credentials.ini
cd config.dist.ini config.ini
```

You'll want to fill in the copied INI files with your own values, which will be described in the **Setup** link below.

# Setup

This repository borrows greatly from the implementation at 
[Flynsarmy](https://www.flynsarmy.com/2015/12/setting-up-dynamic-dns-to-your-home-with-route-53/). Instead of the PHP
script detailed on that page, you can use this repository.

# Usage

Invoke in any number of ways! All assuming you are in the root project directory:

```bash
../path/to/composer execute
```

```bash
./cli
```

```bash
php cli
```

The Flynsarmy article describes running this via an HTTP call. This repository can be executed via the commandline, and
it is recommended to do that for security's sake.

# Recommendations

You should consider scheduling a job to run this on a regular basis. I'm personally using a Linux server with a 
`crontab` job running once per minute.

# Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.
