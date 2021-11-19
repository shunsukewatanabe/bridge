package chain

import (
	"math/big"
)

type SubstrateConfig struct {
	GeneralChainConfig GeneralChainConfig
	StartBlock         *big.Int
	UseExtendedCall    bool
}

type RawSubstrateConfig struct {
	GeneralChainConfig `mapstructure:",squash"`
	StartBlock         int64 `mapstructure:"startBlock"`
	UseExtendedCall    bool  `mapstructure:"useExtendedCall"`
}

func (c *RawSubstrateConfig) ParseConfig() *SubstrateConfig {
	c.GeneralChainConfig.ParseConfig()

	config := &SubstrateConfig{
		GeneralChainConfig: c.GeneralChainConfig,
		StartBlock:         big.NewInt(c.StartBlock),
		UseExtendedCall:    c.UseExtendedCall,
	}
	return config
}
