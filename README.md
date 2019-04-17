# Tutu_search_bleak Service

This is the Tutu_search_bleak service

Generated with

```
micro new tutu_search_bleak/ --namespace=go.micro --alias=tutu_search_bleak --type=srv
```

## Getting Started

- [Configuration](#configuration)
- [Dependencies](#dependencies)
- [Usage](#usage)

## Configuration

- FQDN: go.micro.srv.tutu_search_bleak
- Type: srv
- Alias: tutu_search_bleak

## Dependencies

Micro services depend on service discovery. The default is multicast DNS, a zeroconf system.

In the event you need a resilient multi-host setup we recommend consul.

```
# install consul
brew install consul

# run consul
consul agent -dev
```

## Usage

A Makefile is included for convenience

Build the binary

```
make build
```

Run the service
```
./tutu_search_bleak-srv
```

Build a docker image
```
make docker
```