# dns-drain 
[![Build Status](https://travis-ci.org/czerwonk/dns-drain.svg)](https://travis-ci.org/czerwonk/dns-drain)
[![Go Report Card](https://goreportcard.com/badge/github.com/czerwonk/dns-drain)](https://goreportcard.com/report/github.com/czerwonk/dns-drain)

Drain by removing/replacing IP/net from DNS records with ease

## Remarks
This tool uses GO modules and requires go 1.18+ to build.

## Usage

Drain IP 1.2.3.4 in project api-project-xxx by removing IP from records
```
$ dns-drainctl gcloud --project api-project-xxx drain -f drain.json 1.2.3.4/32
```

Drain IP 1.2.3.4 in project api-project-xxx by replacing IP with 1.2.3.5
```
$ dns-drainctl gcloud --project api-project-xxx drain 1.2.3.4/32 -f drain.json --replace-by 1.2.3.5
```

Undrain by using json file written in drain process
```
$ dns-drainctl gcloud --project api-project-xxx undrain -f drain.json
```

## Supported providers
* Google Cloud DNS

## Future plans
* support for more providers

## License
(c) Daniel Czerwonk, 2016. Licensed under [MIT](LICENSE) license.
