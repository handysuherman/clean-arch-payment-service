package config

type Monitoring struct {
	Probes *Probes `mapstructure:"probes"`
	Jaeger *Jaeger `mapstructure:"jaeger"`
}

type Probes struct {
	ReadinessPath string      `mapstructure:"readinessPath"`
	LivenessPath  string      `mapstructure:"livenessPath"`
	CheckInterval int         `mapstructure:"checkInterval"`
	Port          string      `mapstructure:"port"`
	Prof          string      `mapstructure:"pprof"`
	Prometheus    *Prometheus `mapstructure:"prometheus"`
}

type Prometheus struct {
	Port string `mapstructure:"port"`
	Path string `mapstructure:"path"`
}

type Jaeger struct {
	HostPort string `mapstructure:"hostPort"`
	Enable   bool   `mapstructure:"enable"`
	Logspan  bool   `mapstructure:"logSpan"`
}
