// Copyright 2021 ChainSafe Systems
// SPDX-License-Identifier: LGPL-3.0-only

package evm

import (
	"fmt"
	"math/big"

	"github.com/ChainSafe/chainbridge-core/blockstore"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmclient"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmgaspricer"
	"github.com/ChainSafe/chainbridge-core/chains/evm/evmtransaction"
	"github.com/ChainSafe/chainbridge-core/chains/evm/listener"
	"github.com/ChainSafe/chainbridge-core/chains/evm/voter"
	"github.com/ChainSafe/chainbridge-core/config/chain"
	"github.com/ChainSafe/chainbridge-core/relayer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
)

type EventListener interface {
	ListenToEvents(startBlock *big.Int, domainID uint8, kvrw blockstore.KeyValueWriter, stopChn <-chan struct{}, errChn chan<- error) <-chan *relayer.Message
}

type ProposalVoter interface {
	VoteProposal(message *relayer.Message) error
}

// EVMChain is struct that aggregates all data required for
type EVMChain struct {
	listener EventListener // Rename
	writer   ProposalVoter
	kvdb     blockstore.KeyValueReaderWriter
	config   *chain.SharedEVMConfig
}

func SetupEVMChain(config *evmclient.EVMConfig, db blockstore.KeyValueReaderWriter) (*EVMChain, error) {
	client := evmclient.NewEVMClient()
	err := client.Configurate(config)
	if err != nil {
		return nil, err
	}

	if config.SharedEVMConfig.GeneralChainConfig.LatestBlock {
		latestBlock, err := client.LatestBlock()
		if err != nil {
			return nil, err
		}

		config.SharedEVMConfig.StartBlock = latestBlock
	}

	eventHandler := listener.NewETHEventHandler(common.HexToAddress(config.SharedEVMConfig.Bridge), client)
	eventHandler.RegisterEventHandler(config.SharedEVMConfig.Erc20Handler, listener.Erc20EventHandler)
	eventHandler.RegisterEventHandler(config.SharedEVMConfig.GenericHandler, listener.GenericEventHandler)
	evm1Listener := listener.NewEVMListener(client, eventHandler, common.HexToAddress(config.SharedEVMConfig.Bridge))

	mh := voter.NewEVMMessageHandler(client, common.HexToAddress(config.SharedEVMConfig.Bridge))
	mh.RegisterMessageHandler(common.HexToAddress(config.SharedEVMConfig.Erc20Handler), voter.ERC20MessageHandler)
	mh.RegisterMessageHandler(common.HexToAddress(config.SharedEVMConfig.GenericHandler), voter.GenericMessageHandler)
	evmVoter := voter.NewVoter(mh, client, evmtransaction.NewTransaction, evmgaspricer.NewLondonGasPriceClient(client, nil))

	return NewEVMChain(evm1Listener, evmVoter, db, &config.SharedEVMConfig), nil
}

func NewEVMChain(dr EventListener, writer ProposalVoter, kvdb blockstore.KeyValueReaderWriter, config *chain.SharedEVMConfig) *EVMChain {
	return &EVMChain{listener: dr, writer: writer, kvdb: kvdb, config: config}
}

// PollEvents is the goroutine that polls blocks and searches Deposit events in them.
// Events are then sent to eventsChan.
func (c *EVMChain) PollEvents(stop <-chan struct{}, sysErr chan<- error, eventsChan chan *relayer.Message) {
	log.Info().Msg("Polling Blocks...")

	block, err := blockstore.GetStartingBlock(
		c.kvdb,
		*c.config.GeneralChainConfig.Id,
		c.config.StartBlock,
		c.config.GeneralChainConfig.FreshStart,
	)
	if err != nil {
		sysErr <- fmt.Errorf("error %w on getting last stored block", err)
		return
	}

	ech := c.listener.ListenToEvents(block, *c.config.GeneralChainConfig.Id, c.kvdb, stop, sysErr)
	for {
		select {
		case <-stop:
			return
		case newEvent := <-ech:
			// Here we can place middlewares for custom logic?
			eventsChan <- newEvent
			continue
		}
	}
}

func (c *EVMChain) Write(msg *relayer.Message) error {
	return c.writer.VoteProposal(msg)
}

func (c *EVMChain) DomainID() uint8 {
	return *c.config.GeneralChainConfig.Id
}
