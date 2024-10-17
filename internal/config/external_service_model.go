package config

type External struct {
	PaymentGateway *ExtSvc `mapstructure:"payment_gateway"`
}

type ExtSvc struct {
	ID   string `mapstructure:"id"`
	Name string `mapstructure:"name"`
}
