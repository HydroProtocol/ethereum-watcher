module github.com/HydroProtocol/nights-watch

go 1.12

require (
	github.com/HydroProtocol/hydro-sdk-backend v0.0.39
	github.com/jarcoal/httpmock v1.0.4 // indirect
	github.com/labstack/gommon v0.2.8
	github.com/onrik/ethrpc v0.0.0-20190305112807-6b8e9c0e9a8f
	github.com/shopspring/decimal v0.0.0-20180709203117-cd690d0c9e24
	github.com/sirupsen/logrus v1.4.1
)

// replace github.com/HydroProtocol/hydro-sdk-backend => ../hydro-sdk-backend
