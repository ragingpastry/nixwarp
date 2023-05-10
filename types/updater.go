package types

type UpdateConfig struct {
	PkgConfig UpdatePkgConfig
	NodeConfig UpdateNodeConfig
}

type UpdatePkgConfig struct {
	PkgDir string
}

type UpdateNodeConfig struct {
	AllNodes bool
	Reboot bool
}