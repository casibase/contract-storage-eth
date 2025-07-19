# contract-storage-eth
A simple Ethereum smart contract for storing key-value data with event logging.

## Contract Overview

The `SaveContract` provides functionality to store structured data (key, field, value) and emit events when data is saved. It supports two methods for saving data:
- Save using a `DataItem` struct
- Save using individual string parameters

## Prerequisites

- Go 1.23+
- Local Ethereum development environment (Geth, Ganache, etc.)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/casibase/contract-storage-eth.git
cd contract-storage-eth
```

## Deployment

### Prerequisites for Deployment

1. Start a local Ethereum node using Geth:
```bash
# Start Geth
geth --dev --http --http.api eth,web3,net --datadir ./data --dev --http.addr 0.0.0.0 --http.corsdomain "*"
```

Alternatively, you can use other development environments like ganache

### Method 1: Deploy using Go

1. **Compile the contract to get bytecode and ABI**:
```bash
solc --bin --abi Storage.sol -o build/
```

3. **Run the deployment**:

2. **Edit `config.yaml` to set your JSON-RPC endpoint and private key**:

```yaml
# config.yaml
ethereum:
  # Ethereum node connection URL
  rpc_url: "http://127.0.0.1:8545"
  
  # Private key (without 0x prefix)
  # Note: In production, use environment variables or encrypted storage
  # This is a common test private key, corresponding to address: 0xf39Fd6e51aad88F6F4ce6aB8827279cffFb92266
  # In Geth development mode, this address usually has test ETH
  private_key: "YOUR_PRIVATE_KEY_HERE"
```

Replace `"YOUR_PRIVATE_KEY_HERE"` with your Ethereum account's private key.

Then run the deployment script:

```bash
go run deploy.go
```

The Go script will:
- Connect to your local Ethereum node at `http://127.0.0.1:8545` (Change if needed in `config.yaml`)
- Deploy the contract
- Test the contract by calling the `save` function

### Method 2: Deploy using Remix IDE

1. **Start Geth in development mode** (as shown above)

2. **Open Remix IDE** in your browser: https://remix.ethereum.org

3. **Create new file** in Remix:
    - Click "Create New File" in the file explorer
    - Name it `Storage.sol`
    - Copy and paste the contract code from your local `Storage.sol` file

4. **Compile the contract**:
    - Go to the "Solidity Compiler" tab
    - Select compiler version `0.8.0` or higher
    - Click "Compile Storage.sol"

5. **Connect to local Geth**:
    - Go to the "Deploy & Run Transactions" tab
    - In "Environment", select "Injected Provider - MetaMask" or "External Http Provider"
    - If using External Http Provider, enter: `http://127.0.0.1:8545`

6. **Deploy the contract**:
    - Make sure `SaveContract` is selected in the contract dropdown
    - Click "Deploy"
    - Confirm the transaction
    - Get the contract address from the transaction receipt

7. **Interact with deployed contract**:
    - The deployed contract will appear in the "Deployed Contracts" section
    - You can call functions and view transaction results directly in Remix

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
