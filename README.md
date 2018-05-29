# CF CLI plugin to enable TLS for selected on-demand services

Until BOSH implements [automatic TLS certificate configuration](https://github.com/cloudfoundry/bosh-notes/blob/master/proposals/bosh-dns-cert-sans.md)
the way to enable TLS for on-demand services (that support TLS in the first place) is:
1. `cf create-service`
1. `cf create-service-key` (the key is need to get the address(es) of the ODB instance
1. `cf update-service -c` (update the service with the addresses for TLS
1. `cf delete-service-key`

This CF CLI plugin automates these steps. It provides two commands:

## cf create-service-with-tls
For new service instance, you can automate all four steps. Just run `cf create-service-with-tls` with all the arguments you'd pass to `cf create-service`. For example:
```
cf create-service-with-tls p.mysql db-small mydb
```
This plugin will first execute `cf create-service` and then immediately `cf enable-tls` so you'll get a ready to use service with TLS already enabled.

## cf enable-tls
For existing service instances, you can use `cf enable-tls` to configure TLS. It will perform steps 2-4 for you automatically. For example:
```
cf enable-tls mydb
```

# Compatibility
Currently this plugin supports the following services:
- MySQL for PCF 2.3+
- RabbitMQ for PCF 1.13+

# Installation
```
go build
cf install-plugin cf-cli-enable-tls-plugin
```
