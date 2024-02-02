package promcollectr

var (
	default_port    = "9150"
	default_path    = "/metrics"
	default_cfgPath = "config/exporter"
)

type Option func(*PromcollectrComponent)

func defaultOption() PromcollectrComponent {
	return PromcollectrComponent{
		ready:      true,
		Port:       default_port,
		Path:       default_path,
		CfgPath:    default_cfgPath,
		ServerName: "",
	}
}

func WithPort(port string) Option {
	return func(pm *PromcollectrComponent) {
		pm.Port = port
	}
}

func WithServerName(serverName string) Option {
	return func(pm *PromcollectrComponent) {
		pm.ServerName = serverName
	}
}

func WithPath(path string) Option {
	return func(pm *PromcollectrComponent) {
		pm.Path = path
	}
}

func WithCfgPath(cfgPath string) Option {
	return func(pm *PromcollectrComponent) {
		pm.CfgPath = cfgPath
	}
}
