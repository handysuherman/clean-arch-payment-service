package config

type ServicesKey string

const (
	SERVICES ServicesKey = "services"
)

type Services struct {
	Internal *Internal `mapstructure:"internal"`
	External *External `mapstructure:"external"`
}
