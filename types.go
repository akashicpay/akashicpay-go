package akashicpay

import "time"

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
	Id             string         `json:"id"`                     // Internal Id. Can be ignored
	RequestedValue RequestedValue `json:"requestedValue"`         // Requested value, amount and currency
	ExchangeRate   string         `json:"exchangeRate,omitempty"` // What the received currency is worth in the requested currency
}

type UserInfo struct {
	Identity   string `json:"identity"`   // Akashic address
	WalletType string `json:"walletType"` // Internal. Can be ignored
}

type ITransaction struct {
	FromAddress    string            `json:"fromAddress"`
	ToAddress      string            `json:"toAddress"`
	Layer          TransactionLayer  `json:"layer"`       // Transaction-layer: L1 or L2
	InitiatedAt    string            `json:"initiatedAt"` // Date in ISO8601 format
	ConfirmedAt    string            `json:"confirmedAt"` // Confirmed Date in ISO8601 format
	Amount         string            `json:"amount"`
	CoinSymbol     NetworkSymbol     `json:"coinSymbol"` // Network (L1) of transaction
	Status         TransactionStatus `json:"status"`
	TxHash         string            `json:"txHash,omitempty"`         // Network's hash if L1. Not present for L2
	FeesPaid       string            `json:"feesPaid,omitempty"`       // Gas Fee paid on network. Not present for L2
	L2TxnHash      string            `json:"l2TxnHash,omitempty"`      // Akashic Transaction Hash. For both L1 and L2
	TokenSymbol    TokenSymbol       `json:"tokenSymbol,omitempty"`    // Present only if token-transaction
	InternalFee    InternalFee       `json:"internalFee"`              // Akshic Fee
	Identifier     string            `json:"identifier,omitempty"`     // User-identifier for deposits
	ReferenceId    string            `json:"referenceId,omitempty"`    // Reference-Id to identify a deposit
	DepositRequest DepositRequest    `json:"depositRequest"`           // If a specific value was requested for a deposit, this is included
	ReceiverInfo   UserInfo          `json:"receiverInfo,omitempty"`   // Information about receiving user, for a deposit
	SenderInfo     UserInfo          `json:"senderInfo,omitempty"`     // Information about receiving user, for a withdrawal
	FeeIsDelegated bool              `json:"feeIsDelegated,omitempty"` // Whether network-fee was delegated to a token-fee
}

type IGetTransactions struct {
	Page                  int               // Page, for pagination
	Limit                 int               // Limit for pagination, only accepts 10, 25, 50, or 100
	StartDate             time.Time         // To only include transactions after this time
	EndDate               time.Time         // To only include transactions before this time
	Layer                 TransactionLayer  // Transaction layer: 'L1' or 'L2'
	Status                TransactionStatus // Transaction status: Pending, Confirmed, or Failed
	TransactionType       TransactionType   // Transaction type: Deposit or Withdrawal (Payout)
	HideSmallTransactions bool              // Excludes transactions below 1 USD in value
}

type IDepositAddress struct {
	Address           string // Address (L1) that can be transferred to
	Identifier        string // userId or similar which will be identified with deposits to this address
	ReferenceId       string
	RequestedAmount   string
	RequestedCurrency Currency
	Network           NetworkSymbol
	Token             TokenSymbol
	ExchangeRate      string // Exchange rate of requestedCurrency vs deposit currency
	Amount            string
	Expires           string
	MarkupPercentage  string // Markup percentage to be applied to the exchange rate
}

type iGetByOwnerAndIdentifierResponse struct {
	Address            string `json:"address,omitempty"`
	UnassignedLedgerId string `json:"unassignedLedgerId,omitempty"`
}
type iRequestedValue struct {
	Currency Currency `json:"currency,omitempty"`
	Amount   string   `json:"amount,omitempty"`
}
type iCreateDepositOrderResponse struct {
	Id                string        `json:"id"`
	ReferenceId       string        `json:"referenceId,omitempty"`
	Identifier        string        `json:"identifier"`
	ToAddress         string        `json:"toAddress,omitempty"`
	CoinSymbol        NetworkSymbol `json:"coinSymbol,omitempty"`
	TokenSymbol       TokenSymbol   `json:"tokenSymbol,omitempty"`
	RequestedAmount   string        `json:"requestedAmount,omitempty"`
	RequestedCurrency Currency      `json:"requestedCurrency,omitempty"`
	Amount            string        `json:"amount,omitempty"`
	ExchangeRate      string        `json:"exchangeRate,omitempty"`
	Expires           string        `json:"expires"`
	MarkupPercentage  string        `json:"markupPercentage,omitempty"`
}

type iCreateDepositOrder struct {
	Identity         string           `json:"identity"`
	Expires          int64            `json:"expires"`
	ReferenceId      string           `json:"referenceId"`
	Identifier       string           `json:"identifier"`
	ToAddress        string           `json:"toAddress,omitempty"`
	CoinSymbol       NetworkSymbol    `json:"coinSymbol,omitempty"`
	TokenSymbol      TokenSymbol      `json:"tokenSymbol,omitempty"`
	RequestedValue   *iRequestedValue `json:"requestedValue,omitempty"`
	Signature        string           `json:"signature,omitempty"`
	MarkupPercentage string           `json:"markupPercentage,omitempty"`
}
type iKeyByOwnerAndIdentifierResponse struct {
	CoinSymbol NetworkSymbol `json:"coinSymbol,omitempty"`
	Address    string        `json:"address,omitempty"`
}
