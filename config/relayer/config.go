package relayer

import (
	"fmt"

	"github.com/spf13/viper"
)

type RelayerConfig struct {
	PrometheusEndpoint string `mapstructure:"prometheusEndpoint"`
	PrometheusPort     uint64 `mapstructure:"prometheusPort"`
}

func (c *RelayerConfig) setDefaultValues() {
	viper.SetDefault("PrometheusEndpoint", "/metrics")
	viper.SetDefault("PrometheusPort", 2112)
}

func (c *RelayerConfig) Validate() error {
	c.setDefaultValues()

	if c.PrometheusPort < 1 || c.PrometheusPort > 65535 {
		return fmt.Errorf(`PrometheusPort outside of valid range of 1 - 65535`)
	}

	return nil
}
