# dns-drain [![Build Status](https://travis-ci.org/czerwonk/dns-drain.svg)][travis]
Drain by removing IP/net from records with ease

# Install
```
go get github.com/czerwonk/dns-drain
```
# Usage
Drain IP 1.2.3.4 in project api-project-xxx

```
dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32
```
IP has to be in CIDR notation at the moment

# Supported providers
* Google Cloud DNS

# Future plans
* support for more providers
* undrain functionality

[travis]: https://travis-ci.org/czerwonk/dns-drain
