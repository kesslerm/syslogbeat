// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

type Config struct {
	UDPAddr string `config:"udpaddr"`
	Location string `config:"location"`
}

var DefaultConfig = Config{
	UDPAddr: "0.0.0.0:1514",
	Location: "UTC",
}
