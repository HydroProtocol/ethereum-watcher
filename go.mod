module github.com/HydroProtocol/nights-watch

go 1.12

require (
	github.com/HydroProtocol/hydro-sdk-backend v0.0.38
	github.com/alecthomas/template v0.0.0-20160405071501-a0175ee3bccc // indirect
	github.com/alecthomas/units v0.0.0-20151022065526-2efee857e7cf // indirect
	github.com/go-redis/redis v6.15.2+incompatible
	github.com/jarcoal/httpmock v1.0.4 // indirect
	github.com/onrik/ethrpc v0.0.0-20190305112807-6b8e9c0e9a8f
	github.com/onsi/ginkgo v1.8.0 // indirect
	github.com/onsi/gomega v1.5.0 // indirect
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/common v0.0.0-20181126121408-4724e9255275
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/sirupsen/logrus v1.4.1
	github.com/stretchr/testify v1.3.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6 // indirect
)

replace github.com/HydroProtocol/hydro-sdk-backend => ../hydro-sdk-backend
