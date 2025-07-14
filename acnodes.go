package akashicpay

type ACNode struct {
	Minigate string
	Node     string
}

var ACNodes = map[string]ACNode{
	"Singapore1": {
		Minigate: "https://sg1-minigate.akashicchain.com/",
		Node:     "https://sg1.akashicchain.com/",
	},
	"Singapore2": {
		Minigate: "https://sg2-minigate.akashicchain.com/",
		Node:     "https://sg2.akashicchain.com/",
	},
	"HongKong1": {
		Minigate: "https://hk1-minigate.akashicchain.com/",
		Node:     "https://hk1.akashicchain.com/",
	},
	"HongKong2": {
		Minigate: "https://hk2-minigate.akashicchain.com/",
		Node:     "https://hk2.akashicchain.com/",
	},
	"Japan1": {
		Minigate: "https://jp1-minigate.akashicchain.com/",
		Node:     "https://jp1.akashicchain.com/",
	},
	"Japan2": {
		Minigate: "https://jp2-minigate.akashicchain.com/",
		Node:     "https://jp2.akashicchain.com/",
	},
}

var ACDevNodes = map[string]ACNode{
	"Singapore1": {
		Minigate: "https://sg1-minigate.testnet.akashicchain.com/",
		Node:     "https://sg1.testnet.akashicchain.com/",
	},
	"Singapore2": {
		Minigate: "https://sg2-minigate.testnet.akashicchain.com/",
		Node:     "https://sg2.testnet.akashicchain.com/",
	},
	"Japan1": {
		Minigate: "https://jp1-minigate.testnet.akashicchain.com/",
		Node:     "https://jp1.testnet.akashicchain.com/",
	},
	"Japan2": {
		Minigate: "https://jp2-minigate.testnet.akashicchain.com/",
		Node:     "https://jp2.testnet.akashicchain.com/",
	},
	"HongKong1": {
		Minigate: "https://hk1-minigate.testnet.akashicchain.com/",
		Node:     "https://hk1.testnet.akashicchain.com/",
	},
	"HongKong2": {
		Minigate: "https://hk2-minigate.testnet.akashicchain.com/",
		Node:     "https://hk2.testnet.akashicchain.com/",
	},
}
