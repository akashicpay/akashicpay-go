package akashicpay

type token struct {
	Decimal  int
	Symbol   TokenSymbol
	Contract string
}
type networkInfo struct {
	AddressRegex  string
	NativeDecimal int
	Tokens        []token
}

var networkDictionary = map[NetworkSymbol]networkInfo{
	Ethereum_Mainnet: {
		AddressRegex:  `^0x[A-Fa-f\d]{40}$`,
		NativeDecimal: 18,
		Tokens: []token{
			{
				Decimal:  6,
				Symbol:   USDT,
				Contract: "0xdac17f958d2ee523a2206206994597c13d831ec7",
			},
			{
				Decimal:  6,
				Symbol:   USDC,
				Contract: "0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48",
			},
		},
	},
	Ethereum_Sepolia: {
		AddressRegex:  `^0x[A-Fa-f\d]{40}$`,
		NativeDecimal: 18,
		Tokens: []token{
			{
				Decimal:  6,
				Symbol:   USDT,
				Contract: "0xa62be7ec09f56a813f654a9ac1aa6d29d96f604e",
			},
			{
				Decimal:  6,
				Symbol:   USDC,
				Contract: "0x92ac12b566954e3d0e966cca7c9ddc44ca80ae29",
			},
		},
	},
	Binance_Smart_Chain_Mainnet: {
		AddressRegex:  `^0x[A-Fa-f\d]{40}$`,
		NativeDecimal: 18,
		Tokens: []token{
			{
				Decimal:  18,
				Symbol:   USDT,
				Contract: "0x55d398326f99059ff775485246999027b3197955",
			},
			{
				Decimal:  18,
				Symbol:   USDC,
				Contract: "0x8ac76a51cc950d9822d68b83fe1ad97b32cd580d",
			},
		},
	},
	Binance_Smart_Chain_Testnet: {
		AddressRegex:  `^0x[A-Fa-f\d]{40}$`,
		NativeDecimal: 18,
		Tokens: []token{
			{
				Decimal:  18,
				Symbol:   USDT,
				Contract: "0xa62be7ec09f56a813f654a9ac1aa6d29d96f604e",
			},
			{
				Decimal:  18,
				Symbol:   USDC,
				Contract: "0x9114eb6b5d281ae405c23082cf56043dac280cba",
			},
		},
	},
	Tron: {
		AddressRegex:  `^T[A-Za-z1-9]{33}$`,
		NativeDecimal: 6,
		Tokens: []token{
			{
				Decimal:  6,
				Symbol:   USDT,
				Contract: "TR7NHqjeKQxGTCi8q8ZY4pL8otSzgjLj6t",
			},
		},
	},
	Tron_Shasta: {
		AddressRegex:  `^T[A-Za-z1-9]{33}$`,
		NativeDecimal: 6,
		Tokens: []token{
			{
				Decimal:  6,
				Symbol:   USDT,
				Contract: "TG3XXyExBkPp9nzdajDZsozEu4BkaSJozs",
			},
		},
	},
	Solana: {
		AddressRegex:  `^[1-9A-HJ-NP-Za-km-z]{32,44}$`,
		NativeDecimal: 9,
		Tokens: []token{
			{
				Decimal:  6,
				Symbol:   USDT,
				Contract: "Es9vMFrzaCERmJfrF4H2FYD4KCoNkY11McCe8BenwNYB",
			},
			{
				Decimal:  6,
				Symbol:   USDC,
				Contract: "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v",
			},
		},
	},
	Solana_Devnet: {
		AddressRegex:  `^[1-9A-HJ-NP-Za-km-z]{32,44}$`,
		NativeDecimal: 9,
		Tokens: []token{
			{
				Decimal:  6,
				Symbol:   USDT,
				Contract: "6krZNyyrvgN1XdvLVZwoAY4UvZxgiYVLtJwLYt43GHym",
			},
			{
				Decimal:  6,
				Symbol:   USDC,
				Contract: "7gZkdXQcNzfw4eDJvgN4XuPxBnsf2AyRnjga4XQ7ber8",
			},
		},
	},
}

// NonEthEvmNetworks represents non-Ethereum EVM networks
var NonEthEvmNetworks = []NetworkSymbol{Binance_Smart_Chain_Mainnet, Binance_Smart_Chain_Testnet}
