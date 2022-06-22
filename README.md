# About

**aws-route53-ddns-client** is an AWS Route 53 DDNS client written in Go. If you
can run Go, need a DDNS client, and are willing to use AWS Route 53, then this may
work for you.

## Installation

The recommended installation method is via `git`:

```shell
git clone https://github.com/rfpludwick/aws-route53-ddns-client
```

You'll want to copy the following file in the `conf/` directory:

```shell
cd conf
cp configuration.dist.yaml configuration.yaml
```

You'll want to fill in the copied YAML file with your own values.

## Setup

Make sure you have Go installed and then:

```shell
go build .
```

## Usage

If you've built the executable, then just execute it:

```shell
./aws-route53-ddns-client
```

## Recommendations

You should consider scheduling a job to run this on a regular basis. I've gone
off the deep end and have this running as a `CronJob` in Kubernetes...

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for more details.
