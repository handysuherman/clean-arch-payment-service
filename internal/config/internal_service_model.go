package config

import "time"

type Internal struct {
	ID                 string              `mapstructure:"id"`
	Name               string              `mapstructure:"name"`
	DNS                string              `mapstructure:"dns"`
	LogLevel           string              `mapstructure:"logLevel"`
	Environment        string              `mapstructure:"environment"`
	EnableTLS          bool                `mapstructure:"enableTLS"`
	Addr               string              `mapstructure:"addr"`
	Port               int                 `mapstructure:"port"`
	OperationTimeout   time.Duration       `mapstructure:"operationTimeout"`
	PlatformKeys       *PlatformKeys       `mapstructure:"platformKeys"`
	PaymentGatewayKeys *PaymentGatewayKeys `mapstructure:"paymentGatewayKeys"`
}

type PlatformKeys struct {
	Mobile  string `mapstructure:"mobile"`
	Website string `mapstructure:"website"`
}

type PaymentGatewayKeys struct {
	Development string `mapstructure:"development"`
	Production  string `mapstructure:"production"`
}
