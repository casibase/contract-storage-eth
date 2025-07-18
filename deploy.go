// Copyright 2025 contract-storage-eth Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"gopkg.in/yaml.v2"
)

// Config structure for deployment configuration
type Config struct {
	Ethereum struct {
		RpcURL     string `yaml:"rpc_url"`
		PrivateKey string `yaml:"private_key"`
		ChainID    int64  `yaml:"chain_id"`
		GasLimit   uint64 `yaml:"gas_limit"`
	} `yaml:"ethereum"`
	Build struct {
		Directory    string `yaml:"directory"`
		ContractName string `yaml:"contract_name"`
	} `yaml:"build"`
	Test struct {
		Enable    bool   `yaml:"enable"`
		TestKey   string `yaml:"test_key"`
		TestField string `yaml:"test_field"`
		TestValue string `yaml:"test_value"`
	} `yaml:"test"`
}

func main() {
	fmt.Println("Starting contract deployment...")

	// Load configuration file
	config, err := loadConfig("config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Connect to Ethereum node
	client, err := ethclient.Dial(config.Ethereum.RpcURL)
	if err != nil {
		log.Fatal("Failed to connect to Ethereum node:", err)
	}
	defer client.Close()
	fmt.Printf("Connected to Ethereum node: %s\n", config.Ethereum.RpcURL)

	// Load private key
	privateKey, err := crypto.HexToECDSA(config.Ethereum.PrivateKey)
	if err != nil {
		log.Fatal("Failed to load private key:", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	fmt.Printf("Deploying from address: %s\n", fromAddress.Hex())

	// Read contract bytecode
	bytecodeFile := filepath.Join(config.Build.Directory, config.Build.ContractName+".bin")
	bytecodeBytes, err := os.ReadFile(bytecodeFile)
	if err != nil {
		log.Fatal("Failed to read bytecode file:", err)
	}
	bytecode := strings.TrimSpace(string(bytecodeBytes))
	fmt.Printf("Loaded bytecode from: %s\n", bytecodeFile)

	// Read contract ABI
	abiFile := filepath.Join(config.Build.Directory, config.Build.ContractName+".abi")
	abiBytes, err := os.ReadFile(abiFile)
	if err != nil {
		log.Fatal("Failed to read ABI file:", err)
	}
	abiString := strings.TrimSpace(string(abiBytes))
	fmt.Printf("Loaded ABI from: %s\n", abiFile)

	// Parse ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiString))
	if err != nil {
		log.Fatal("Failed to parse ABI:", err)
	}

	// Get nonce
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		log.Fatal("Failed to get nonce:", err)
	}

	// Get gas price
	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		log.Fatal("Failed to get gas price:", err)
	}

	// Get chain ID
	chainID, err := client.ChainID(context.Background())
	if err != nil {
		log.Fatal("Failed to get chain ID:", err)
	}

	// Create auth object
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Fatal("Failed to create auth:", err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = config.Ethereum.GasLimit
	auth.GasPrice = gasPrice

	fmt.Printf("Gas price: %s wei\n", gasPrice.String())
	fmt.Printf("Gas limit: %d\n", auth.GasLimit)

	// Deploy contract
	fmt.Println("Deploying contract...")
	bytecodeData := common.FromHex(bytecode)
	address, tx, _, err := bind.DeployContract(auth, parsedABI, bytecodeData, client)
	if err != nil {
		log.Fatal("Failed to deploy contract:", err)
	}

	fmt.Printf("Transaction sent: %s\n", tx.Hash().Hex())
	fmt.Printf("Contract address: %s\n", address.Hex())

	// Wait for transaction confirmation
	fmt.Println("Waiting for transaction confirmation...")
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Fatal("Failed to wait for transaction:", err)
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Println("Contract deployed successfully!")
		fmt.Printf("Gas used: %d\n", receipt.GasUsed)
		fmt.Printf("Block number: %d\n", receipt.BlockNumber.Uint64())
	} else {
		log.Fatal("Contract deployment failed!")
	}

	// Optional testing
	if config.Test.Enable {
		fmt.Println("\nRunning contract test...")
		testContract(client, address, privateKey, chainID, parsedABI, config)
	}

	fmt.Println("\nDeployment completed!")
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}

func testContract(client *ethclient.Client, contractAddress common.Address, privateKey *ecdsa.PrivateKey, chainID *big.Int, parsedABI abi.ABI, config *Config) {
	// Create contract instance
	contract := bind.NewBoundContract(contractAddress, parsedABI, client, client, client)

	// Create auth object
	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	if err != nil {
		log.Printf("Failed to create auth for testing: %v", err)
		return
	}
	auth.GasLimit = uint64(300000)

	// Call save function
	fmt.Printf("Calling save function with: key=%s, field=%s, value=%s\n",
		config.Test.TestKey, config.Test.TestField, config.Test.TestValue)

	tx, err := contract.Transact(auth, "save", config.Test.TestKey, config.Test.TestField, config.Test.TestValue)
	if err != nil {
		log.Printf("Failed to call save function: %v", err)
		return
	}

	fmt.Printf("Save transaction: %s\n", tx.Hash().Hex())

	// Wait for transaction confirmation
	receipt, err := bind.WaitMined(context.Background(), client, tx)
	if err != nil {
		log.Printf("Failed to wait for save transaction: %v", err)
		return
	}

	if receipt.Status == types.ReceiptStatusSuccessful {
		fmt.Println("Save function called successfully!")
		// Read data back
		var result []interface{}
		err = contract.Call(nil, &result, "data")
		if err != nil {
			log.Printf("Failed to read data: %v", err)
			return
		}

		if len(result) == 3 {
			fmt.Printf("Retrieved data - Key: %s, Field: %s, Value: %s\n",
				result[0].(string), result[1].(string), result[2].(string))
		}

		// Check logs
		for _, log := range receipt.Logs {
			if log.Address == contractAddress {
				logData, err := parsedABI.Unpack("DataSaved", log.Data)
				if err != nil {
					fmt.Printf("Failed to unpack log data: %v", err)
					continue
				}
				if len(logData) == 3 {
					fmt.Printf("Log data - Key: %s, Field: %s, Value: %s\n",
						logData[0].(string), logData[1].(string), logData[2].(string))
				}
			}
		}
	} else {
		fmt.Println("Save function call failed!")
	}
}
