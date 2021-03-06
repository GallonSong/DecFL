package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"

	"github.com/FMorsbach/DecFL/model/chain"
	"github.com/FMorsbach/DecFL/model/chain/ethereum/contract"
	"github.com/FMorsbach/DecFL/model/common"
	"github.com/FMorsbach/dlog"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethCommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

var logger = dlog.New(os.Stderr, "Chain: ", log.LstdFlags, false)

func EnableDebug(b bool) {
	logger.SetDebug(b)
}

type ethereumChain struct {
	client        ethclient.Client
	privateKey    ecdsa.PrivateKey
	publicKey     ecdsa.PublicKey
	publicAddress ethCommon.Address
}

func (c *ethereumChain) waitForTransaction(hash ethCommon.Hash) error {

	// for {
	// 	time.Sleep(5 * time.Second)

	// 	_, isPending, err := c.client.TransactionByHash(context.Background(), hash)
	// 	if err != nil {
	// 		panic(err)
	// 		return err
	// 	} else if !isPending {
	// 		break
	// 	} else {
	// 		logger.Debugln("Still pending transactions, sleeping for 5 second.")
	// 	}
	// }

	return nil
}

func NewEthereum(chainAddress string, key string) (instance chain.Chain, err error) {

	client, err := ethclient.Dial(chainAddress)
	if err != nil {
		return
	}

	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		return
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		err = fmt.Errorf("%s%s", "Cannot assert type", err)
		return
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	return &ethereumChain{
		client:        *client,
		privateKey:    *privateKey,
		publicKey:     *publicKeyECDSA,
		publicAddress: fromAddress,
	}, nil
}

func (c *ethereumChain) DeployModel(configAddress common.StorageAddress, weightsAddress common.StorageAddress, scriptsAddress common.StorageAddress, params common.Hyperparameters) (id common.ModelIdentifier, err error) {

	nonce, err := c.client.PendingNonceAt(context.Background(), c.publicAddress)
	if err != nil {
		return

	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth := bind.NewKeyedTransactor(&(c.privateKey))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice

	address, tx, _, err := contract.DeployContract(
		auth,
		&(c.client),
		string(configAddress),
		string(weightsAddress),
		string(scriptsAddress),
		big.NewInt(int64(params.UpdatesTillAggregation)),
		big.NewInt(int64(params.ConsensusThreshold)),
		big.NewInt(int64(params.Epochs)),
	)
	if err != nil {
		return
	}

	err = c.waitForTransaction(tx.Hash())
	if err != nil {
		return
	}

	id = common.ModelIdentifier(address.Hex())
	logger.Debugf("Deployed contract in transaction %s", tx.Hash().Hex())

	return
}

func (c *ethereumChain) ModelConfigurationAddress(id common.ModelIdentifier) (address common.StorageAddress, err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	value, err := instance.ConfigurationAddress(nil)
	if err != nil {
		return
	}

	address = common.StorageAddress(value)
	return
}

func (c *ethereumChain) GlobalWeightsAddress(id common.ModelIdentifier) (address common.StorageAddress, err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	value, err := instance.WeightsAddress(nil)
	if err != nil {
		return
	}

	address = common.StorageAddress(value)
	return
}

func (c *ethereumChain) ScriptsAddress(id common.ModelIdentifier) (address common.StorageAddress, err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	value, err := instance.ScriptsAddress(nil)
	if err != nil {
		return
	}

	address = common.StorageAddress(value)
	return
}

func (c *ethereumChain) SubmitAggregation(id common.ModelIdentifier, address common.StorageAddress) (err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	nonce, err := c.client.PendingNonceAt(context.Background(), c.publicAddress)
	if err != nil {
		return

	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth := bind.NewKeyedTransactor(&(c.privateKey))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice

	tx, err := instance.SubmitLocalAggregation(auth, string(address))
	if err != nil {
		// panic(err)
		return
	}

	err = c.waitForTransaction(tx.Hash())
	if err != nil {
		return
	}

	logger.Debugf("Wrote local update to chain as tx: %s", tx.Hash().Hex())
	return
}

func (c *ethereumChain) SubmitLocalUpdate(id common.ModelIdentifier, updateAddress common.StorageAddress) (err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	nonce, err := c.client.PendingNonceAt(context.Background(), c.publicAddress)
	if err != nil {
		return

	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth := bind.NewKeyedTransactor(&(c.privateKey))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice

	tx, err := instance.SubmitLocalUpdate(auth, string(updateAddress))
	if err != nil {
		return
	}

	err = c.waitForTransaction(tx.Hash())
	if err != nil {
		return
	}

	logger.Debugf("Wrote local update to chain as tx: %s", tx.Hash().Hex())
	return
}

func (c *ethereumChain) LocalUpdates(id common.ModelIdentifier) (updates []common.Update, err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	count, err := instance.LocalUpdatesCount(nil)
	if err != nil {
		return
	}

	for i := big.NewInt(0); i.Cmp(count) == -1; i.Add(i, big.NewInt(1)) {

		update, err := instance.LocalUpdates(nil, i)
		if err != nil {
			return nil, err
		}
		updates = append(updates, common.Update{
			Trainer: common.TrainerIdentifier(update.Trainer),
			Address: common.StorageAddress(update.StorageAddress),
		})
	}

	return
}

func (c *ethereumChain) ModelEpoch(id common.ModelIdentifier) (epoch int, err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	value, err := instance.CurrentEpoch(nil)
	if err != nil {
		return
	}

	epoch = int(value.Int64())
	return
}

func (c *ethereumChain) State(id common.ModelIdentifier) (state uint8, err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	state, err = instance.State(nil)

	return
}

func (c *ethereumChain) AddTrainer(id common.ModelIdentifier, trainer common.TrainerIdentifier) (err error) {

	instance, err := contract.NewContract(ethCommon.HexToAddress(string(id)), &(c.client))
	if err != nil {
		return
	}

	nonce, err := c.client.PendingNonceAt(context.Background(), c.publicAddress)
	if err != nil {
		return

	}

	gasPrice, err := c.client.SuggestGasPrice(context.Background())
	if err != nil {
		return
	}

	auth := bind.NewKeyedTransactor(&(c.privateKey))
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)      // in wei
	auth.GasLimit = uint64(3000000) // in units
	auth.GasPrice = gasPrice

	tx, err := instance.AddTrainer(auth, ethCommon.Address(trainer))
	if err != nil {
		return
	}

	err = c.waitForTransaction(tx.Hash())
	if err != nil {
		return
	}

	logger.Debugf("Added trainer to contract as tx: %s", tx.Hash().Hex())
	return
}
