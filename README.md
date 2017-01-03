# dns-drain [![Build Status](https://travis-ci.org/czerwonk/dns-drain.svg)][travis]
Drain by removing/replacing IP/net from DNS records with ease

# Install
```
go get github.com/czerwonk/dns-drain
```
# Usage

## Drain IP 1.2.3.4 in project api-project-xxx
```
dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32
``` 

## Undrain by using json file written in drain process
```
dns-drain -undrain -gcloud.project=api-project-xxx
```

# Parameters

Name     | Description
---------|-------------
-ip | define which IPs/nets in records should be matched  (has to be in CIDR notation at the moment)
-dry | simulation only
-zone | apply changes for specific zones only
-skip | skip specific zones
-new_ip | set a replacement IP (IP only)

# Supported providers
* Google Cloud DNS

# Future plans
* support for more providers

[travis]: https://travis-ci.org/czerwonk/dns-drain
