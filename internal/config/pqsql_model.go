package config

type PqSQLKey string

const (
	PQSQL PqSQLKey = "postgresql"
)

type PostgreSQL struct {
	Driver       string `mapstructure:"driver"`
	Source       string `mapstructure:"source"`
	TLsSource    string `mapstructure:"tlsSource"`
	Username     string `mapstructure:"username"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"dbName"`
	Host         string `mapstructure:"host"`
	Port         string `mapstructure:"port"`
	MigrationURL string `mapstructure:"migrationURL"`
	EnableTLS    bool   `mapstructure:"enableTLS"`
	Capath       string `mapstructure:"ca"`
	Certpath     string `mapstructure:"cert"`
	Keypath      string `mapstructure:"key"`
}
