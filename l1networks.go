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
			{
				Decimal:  6,
				Symbol:   tether,
				Contract: "TG3XXyExBkPp9nzdajDZsozEu4BkaSJozs",
			},
		},
	},
}
