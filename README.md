# AkashicPay - Go SDK

A library to interact with the AkashicChain network for Go.

# Installing

Install the package with:

```sh
go get github.com/akashicpay/akashicpay-go
```

# Usage

**Features**

- Send crypto via **Layer 1** and **Layer 2** (Akashic Chain)
- Create wallets for your users into which they can deposit crypto
- Fetch balance and transaction details
- Completely _Web3_: No login or API-key necessary. Just supply your Akashic
  private key, which stays on your server. The SDK signs your transactions with your key and sends them to Akashic Chain.
- Supports **Ethereum** and **Tron**

**Getting Started**

1. Create an account on AkashicLink (Google Chrome Extension or iPhone/Android
   App)
2. Visit [AkashicPay](https://www.akashicpay.com) and connect with AkashicLink.
   Set up the URL(s) you wish to receive callbacks for.
3. Integrate the SDK in your code. Example:

```Go
import (
    "os"

    akashicpay "github.com/akashic/go-sdk"
)
// use whatever secret management tool you prefer to load the private key
// from your AkashicLink account. It should be of the form:
// "0x2d99270559d7702eadd1c5a483d0a795566dc76c18ad9d426c932de41bfb78b7"
apKey := os.Getenv("ApKey")
// this is the address of your AkashicLink account. Of the form "AS1234..."
apL2Address := os.Getenv("ApL2Address")

// in development, you will use our testnet and testnet L1 chains
env := os.Getenv("Environment")
apEnv := akashicpay.Development

if env == "Prod" {
  apEnv = akashicPay.Production
}

// instantiate an SDK instance, ready to use
ap, err := akashicpay.NewAkashicPay(apKey, apL2Address, apEnv, "")
```

AkashicPay is now fully setup and ready to use.

# Testing

You can also use AkashicPay with the AkashicChain Testnet & **Sepolia**
(Ethereum) and **Shasta** (Tron) testnets, useful for local development and
preprod environments.
To do this, follow the same procedure as above but make sure you use _testnet_
versions of AkashicLink and [AkashicPay](https://testnet.akashicpay.com). Make
sure you use the development environment in the SDK:

```Go
import (
	"os"

	akashicpay "github.com/akashic/go-sdk"
)

apKey := os.Getenv("ApTestKey")
apL2Address := os.Getenv("ApTestL2Address")

// in development, you will use our testnet and testnet L1 chains
apEnv := akashicpay.Development

// instantiate an SDK instance, ready to use
ap, err := akashicpay.NewAkashicPay(apKey, apL2Address, apEnv, "")

if err != nil {
// handle error
}

```

You can now create an L1-wallet on a testnet:

```Go
dA, err := ap.GetDepositAddress(akashicpay.Tron_Shasta, "user123", "")
```

## Faucet

During testing and local development, you need cryptocurrency on the testnets to do anything meaningful. Akashic provides a simple faucet where you can request some coins and tokens on Shasta and Sepolia by supplying your L2-address/identity: https://faucet.testnet.akashicchain.com/
If you require further funds, the official Tron Discord provides users with
either 5000 TRX or USDT on Shasta every day.

You can check to see if your balance has increased with:

```Go
balance, err := ap.GetBalance()
// -> [{networkSymbol: 'TRX-SHASTA', balance: '5000'}, ...]
```

# Documentation

For more in-depth documentation describing the SDKs functions in detail,
explanations of terminology, and guides on how to use AkashicPay.com, click [here](https://docs.akashicpay.com/)

# License

This project is licensed under the [MIT](./LICENSE) License
