package config

type Databases struct {
	PostgreSQL *PostgreSQL `mapstructure:"postgresql"`
	Redis      *Redis      `mapstructure:"redis"`
}
