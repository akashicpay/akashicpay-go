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
	TransactionType       TransactionType
	HideSmallTransactions bool
}

type IDepositAddress struct {
	Address           string
	Identifier        string
	ReferenceId       string
	RequestedAmount   string
	RequestedCurrency Currency
	Network           NetworkSymbol
	Token             TokenSymbol
	ExchangeRate      string
	Amount            string
	Expires           string
	MarkupPercentage  string
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
