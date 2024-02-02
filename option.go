package promcollectr

var (
	default_name = "promcollectr"
)

var (
	default_port    = ":9150"
	default_path    = "/metrics"
	default_cfgPath = "config/exporter"
)

type Option func(*PromcollectrServer)

func defaultOption() PromcollectrServer {
	return PromcollectrServer{
		ready:      true,
		name:       default_name,
		Port:       default_port,
		Path:       default_path,
		CfgPath:    default_cfgPath,
		ServerName: "",
	}
}

func WithPort(port string) Option {
	return func(pm *PromcollectrServer) {
		pm.Port = port
	}
}

func WithServerName(serverName string) Option {
	return func(pm *PromcollectrServer) {
		pm.ServerName = serverName
	}
}

func WithPath(path string) Option {
	return func(pm *PromcollectrServer) {
		pm.Path = path
	}
}

func WithCfgPath(cfgPath string) Option {
	return func(pm *PromcollectrServer) {
		pm.CfgPath = cfgPath
	}
}
