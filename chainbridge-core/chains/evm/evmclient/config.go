package evmclient

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shunsukewatanabe/bridge/chainbridge-core/config/chain"
	"github.com/shunsukewatanabe/bridge/chainbridge-core/crypto/secp256k1"
	"github.com/spf13/viper"
)

type EVMConfig struct {
	SharedEVMConfig chain.SharedEVMConfig
	kp              *secp256k1.Keypair
}

type RawEVMConfig struct {
	chain.RawSharedEVMConfig `mapstructure:",squash"`
}

func NewConfig() *EVMConfig {
	return &EVMConfig{}
}

func GetConfig(path string, name string) (*RawEVMConfig, error) {
	config := &RawEVMConfig{}

	viper.AddConfigPath(path)
	viper.SetConfigName(name)
	viper.SetConfigType("json")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read in the config file, error: %w", err)
	}

	if err := viper.Unmarshal(config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config into struct, error: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func ParseConfig(rawConfig *RawEVMConfig) (*EVMConfig, error) {
	cfg, err := rawConfig.RawSharedEVMConfig.ParseConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to parse shared evm config, error: %w", err)
	}

	config := &EVMConfig{
		SharedEVMConfig: *cfg,
	}

	return config, nil
}

func (c *RawEVMConfig) ToJSON(file string) *os.File {
	var (
		newFile *os.File
		err     error
	)

	var raw []byte
	if raw, err = json.Marshal(&c); err != nil {
		fmt.Println("error marshalling json", "err", err)
		os.Exit(1)
	}

	newFile, err = os.Create(file)
	if err != nil {
		fmt.Println("error creating config file", "err", err)
	}
	_, err = newFile.Write(raw)
	if err != nil {
		fmt.Println("error writing to config file", "err", err)
	}

	if err := newFile.Close(); err != nil {
		fmt.Println("failed to unmarshal config into struct", "err", err)
	}
	return newFile
}
