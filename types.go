package akashicpay

import "time"

type Currency string

const (
	// Crypto
	USDT_Currency Currency = "USDT"
	TRX_Currency  Currency = "TRX"
	ETH_Currency  Currency = "ETH"

	// Fiat
	CHF Currency = "CHF"
	CNY Currency = "CNY"
	EUR Currency = "EUR"
	HKD Currency = "HKD"
	IDR Currency = "IDR"
	JPY Currency = "JPY"
	KHR Currency = "KHR"
	KRW Currency = "KRW"
	MYR Currency = "MYR"
	PHP Currency = "PHP"
	SGD Currency = "SGD"
	THB Currency = "THB"
	TWD Currency = "TWD"
	USD Currency = "USD"
	VND Currency = "VND"
)

type TransactionLayer string

const (
	L1 TransactionLayer = "L1Transaction"
	L2 TransactionLayer = "L2Transaction"
)

type TransactionStatus string

const (
	PENDING   TransactionStatus = "Pending"
	CONFIRMED TransactionStatus = "Confirmed"
	FAILED    TransactionStatus = "Failed"
)

type TransactionType string

const (
	DEPOSIT    TransactionType = "Deposit"
	WITHDRAWAL TransactionType = "Withdrawal"
)

type InternalFee struct {
	Deposit  string `json:"deposit,omitempty"`
	Withdraw string `json:"withdraw,omitempty"`
}

type RequestedValue struct {
	Amount   string   `json:"amount"`
	Currency Currency `json:"currency"`
}

type DepositRequest struct {
	Id             string         `json:"id"`
	RequestedValue RequestedValue `json:"requestedValue"`
	ExchangeRate   string         `json:"exchangeRate,omitempty"`
}

type UserInfo struct {
	Identity   string `json:"identity"`
	WalletType string `json:"walletType"`
}

type ITransaction struct {
	FromAddress    string            `json:"fromAddress"`
	ToAddress      string            `json:"toAddress"`
	Layer          TransactionLayer  `json:"layer"`
	InitiatedAt    string            `json:"initiatedAt"`
	ConfirmedAt    string            `json:"confirmedAt"`
	Amount         string            `json:"amount"`
	CoinSymbol     NetworkSymbol     `json:"coinSymbol"`
	Status         TransactionStatus `json:"status"`
	TxHash         string            `json:"txHash,omitempty"`
	FeesPaid       string            `json:"feesPaid,omitempty"`
	L2TxnHash      string            `json:"l2TxnHash,omitempty"`
	TokenSymbol    TokenSymbol       `json:"tokenSymbol,omitempty"`
	InternalFee    InternalFee       `json:"internalFee"`
	Identifier     string            `json:"identifier,omitempty"`
	ReferenceId    string            `json:"referenceId,omitempty"`
	DepositRequest DepositRequest    `json:"depositRequest"`
	ReceiverInfo   UserInfo          `json:"receiverInfo"`
	SenderInfo     UserInfo          `json:"senderInfo"`
	FeeIsDelegated bool              `json:"feeIsDelegated,omitempty"`
}

type IGetTransactions struct {
	Page                  int
	Limit                 int
	StartDate             time.Time
	EndDate               time.Time
	Layer                 TransactionLayer
	Status                TransactionStatus
	Type                  TransactionType
	HideSmallTransactions bool
}
