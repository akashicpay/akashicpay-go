package akashicpay

import "fmt"

type akashicErrorCode string

const (
	akashicErrorCodeTestNetOtkOnboardingFailed akashicErrorCode = "OTK_ONBOARDING_FAILED"
	akashicErrorCodeIncorrectPrivateKeyFormat  akashicErrorCode = "INVALID_PRIVATE_KEY_FORMAT"
	akashicErrorCodeUnknownError               akashicErrorCode = "UNKNOWN_ERROR"
	akashicErrorCodeKeyCreationFailure         akashicErrorCode = "WALLET_CREATION_FAILURE"
	akashicErrorCodeUnHealthyKey               akashicErrorCode = "UNHEALTHY_WALLET"
	akashicErrorCodeAccessDenied               akashicErrorCode = "ACCESS_DENIED"
	akashicErrorCodeL2AddressNotFound          akashicErrorCode = "L2ADDRESS_NOT_FOUND"
	akashicErrorCodeIsNotBp                    akashicErrorCode = "NOT_SIGNED_UP"
	akashicErrorCodeSavingsExceeded            akashicErrorCode = "FUNDS_EXCEEDED"
	akashicErrorCodeAssignmentFailed           akashicErrorCode = "ASSIGNMENT_FAILED"
	akashicErrorCodeNetworkEnvironmentMismatch akashicErrorCode = "NETWORK_ENVIRONMENT_MISMATCH"
)

var akashicErrorDetail = map[akashicErrorCode]string{
	akashicErrorCodeTestNetOtkOnboardingFailed: "failed to setup test-otk. Please try again",
	akashicErrorCodeIncorrectPrivateKeyFormat:  "private Key is not in correct format",
	akashicErrorCodeUnknownError:               "akashic failed with an unknown error. Please try again later",
	akashicErrorCodeKeyCreationFailure:         "failed to generate new wallet. Try again.",
	akashicErrorCodeUnHealthyKey:               "new wallet was not created safely, please re-create",
	akashicErrorCodeAccessDenied:               "unauthorized attempt to access production Akashic Link secrets",
	akashicErrorCodeL2AddressNotFound:          "l2 address not found",
	akashicErrorCodeIsNotBp:                    "please sign up on AkashicPay.com first",
	akashicErrorCodeSavingsExceeded:            "transaction amount exceeds total savings",
	akashicErrorCodeAssignmentFailed:           "failed to assign wallet. Please try again",
	akashicErrorCodeNetworkEnvironmentMismatch: "the L1-network does not match the SDK-environment",
}

type akashicError struct {
	Code    akashicErrorCode
	Details string
}

func (e *akashicError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Details)
}

func newAkashicError(code akashicErrorCode, details string) *akashicError {
	var Details string
	if details == "" {
		Details = akashicErrorDetail[code]
	}
	return &akashicError{
		Code:    code,
		Details: Details,
	}
}
