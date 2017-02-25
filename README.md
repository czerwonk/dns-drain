# dns-drain [![Build Status](https://travis-ci.org/czerwonk/dns-drain.svg)][travis]
Drain by removing/replacing IP/net from DNS records with ease

# Install
```
go get github.com/czerwonk/dns-drain
```
# Usage

Drain IP 1.2.3.4 in project api-project-xxx by removing IP from records
```
dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32
``` 

Drain IP 1.2.3.4 in project api-project-xxx by replacing IP with 1.2.3.5
```
dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32 -new_ip=1.2.3.5
```

Undrain by using json file written in drain process
```
dns-drain -undrain -gcloud.project=api-project-xxx
```

# Parameters

Name        | Description
------------|------------
-ip | defines which IPs/nets in records should be matched
-value | defines which values in records should be matched
-regex | defines which pattern should be applied to match records
-dry | simulation only
-zone | apply changes for specific zones only (regular expression)
-skip | skip specific zones (regular expression)
-new_ip | set a replacement IP (not compatible with -value and -regex)
-new_value | set a replacement value (not compatible with -ip)
-file | input (undrain) or output (drain) file (default: drain.json)

# Supported providers
* Google Cloud DNS

# Future plans
* support for more providers

[travis]: https://travis-ci.org/czerwonk/dns-drain
