package config

var (
	gConfig = &TrustedEngineConfig{}
)

type TrustedEngineConfig struct {
	// Trusted engine
	TrustedClient string `toml:",omitempty"`

	// Grpc Server
	ChainServer string `toml:",omitempty"`
}

func SetupTrustedEngineConfig(clientAddr string, chainServerAddr string) {
	gConfig.TrustedClient = clientAddr
	gConfig.TrustedClient = chainServerAddr
}

func GetConfig() *TrustedEngineConfig {
	return gConfig
}
