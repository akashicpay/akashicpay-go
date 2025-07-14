package akashicpay

import "fmt"

type AkashicErrorCode string

const (
	AkashicErrorCodeTestNetOtkOnboardingFailed AkashicErrorCode = "OTK_ONBOARDING_FAILED"
	AkashicErrorCodeIncorrectPrivateKeyFormat  AkashicErrorCode = "INVALID_PRIVATE_KEY_FORMAT"
	AkashicErrorCodeUnknownError               AkashicErrorCode = "UNKNOWN_ERROR"
	AkashicErrorCodeKeyCreationFailure         AkashicErrorCode = "WALLET_CREATION_FAILURE"
	AkashicErrorCodeUnHealthyKey               AkashicErrorCode = "UNHEALTHY_WALLET"
	AkashicErrorCodeAccessDenied               AkashicErrorCode = "ACCESS_DENIED"
	AkashicErrorCodeL2AddressNotFound          AkashicErrorCode = "L2ADDRESS_NOT_FOUND"
	AkashicErrorCodeIsNotBp                    AkashicErrorCode = "NOT_SIGNED_UP"
	AkashicErrorCodeSavingsExceeded            AkashicErrorCode = "FUNDS_EXCEEDED"
	AkashicErrorCodeAssignmentFailed           AkashicErrorCode = "ASSIGNMENT_FAILED"
	AkashicErrorCodeNetworkEnvironmentMismatch AkashicErrorCode = "NETWORK_ENVIRONMENT_MISMATCH"
)

var AkashicErrorDetail = map[AkashicErrorCode]string{
	AkashicErrorCodeTestNetOtkOnboardingFailed: "failed to setup test-otk. Please try again",
	AkashicErrorCodeIncorrectPrivateKeyFormat:  "private Key is not in correct format",
	AkashicErrorCodeUnknownError:               "akashic failed with an unknown error. Please try again later",
	AkashicErrorCodeKeyCreationFailure:         "failed to generate new wallet. Try again.",
	AkashicErrorCodeUnHealthyKey:               "new wallet was not created safely, please re-create",
	AkashicErrorCodeAccessDenied:               "unauthorized attempt to access production Akashic Link secrets",
	AkashicErrorCodeL2AddressNotFound:          "l2 address not found",
	AkashicErrorCodeIsNotBp:                    "please sign up on AkashicPay.com first",
	AkashicErrorCodeSavingsExceeded:            "transaction amount exceeds total savings",
	AkashicErrorCodeAssignmentFailed:           "failed to assign wallet. Please try again",
	AkashicErrorCodeNetworkEnvironmentMismatch: "the L1-network does not match the SDK-environment",
}

type AkashicError struct {
	Code    AkashicErrorCode
	Details string
}

func (e *AkashicError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Details)
}

func NewAkashicError(code AkashicErrorCode, details string) *AkashicError {
	var Details string
	if details == "" {
		Details = AkashicErrorDetail[code]
	}
	return &AkashicError{
		Code:    code,
		Details: Details,
	}
}
