# dns-drain[![Build Status](https://travis-ci.org/czerwonk/dns-drain.svg)][travis]
DNS/Frontend Drain with ease

# Remarks
this is an early version

# Install
```
go get github.com/czerwonk/dns-drain
```
# Usage
Drain frontend with IP 1.2.3.4 in Project api-project-xxx

```
dns-drain -gcloud.project=api-project-xxx -ip=1.2.3.4/32
```

# Supported providers
* Google Cloud DNS

# Future plans
* support for more providers
* undrain functionality

[travis]: https://travis-ci.org/czerwonk/dns-drain
