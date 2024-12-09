package rri

import (
	"fmt"
	"strings"
	"time"
)

const VerificationInformationTimestampFormat = "2006-01-02T15:04:05-07:00"

// VerificationInformation holds verification information.
type VerificationInformation struct {
	VerificationTimestamp time.Time
	VerificationResult    VerificationResult
	VerificationReference string
	VerificationEvidence  VerificationEvidence
	VerificationMethod    VerificationMethod
	TrustFramework        TrustFramework
	VerifiedClaim         []VerificationClaim
}

func (verificationInformation *VerificationInformation) PutToQueryFields(fields *QueryFieldList) {
	fields.Add(QueryFieldNameEntity, QueryEntityVerificationInformation.String())

	verifiedClaimSlice := make([]string, len(verificationInformation.VerifiedClaim))
	for i := range verificationInformation.VerifiedClaim {
		verifiedClaimSlice[i] = string(verificationInformation.VerifiedClaim[0])
	}

	fields.Add(QueryFieldNameVerifiedClaim, verifiedClaimSlice...)
	fields.Add(QueryFieldNameVerificationResult, string(verificationInformation.VerificationResult))
	fields.Add(QueryFieldNameVerificationReference, verificationInformation.VerificationReference)
	fields.Add(QueryFieldNameVerificationTimestamp, verificationInformation.VerificationTimestamp.Format(VerificationInformationTimestampFormat))
	fields.Add(QueryFieldNameVerificationEvidence, string(verificationInformation.VerificationEvidence))
	fields.Add(QueryFieldNameVerificationMethod, string(verificationInformation.VerificationMethod))
	fields.Add(QueryFieldNameTrustFramework, string(verificationInformation.TrustFramework))
}

type VerificationResult string

const (
	VerificationResultSuccess VerificationResult = "success"
	VerificationResultFailed  VerificationResult = "failed"
)

// ParseVerificationResult parses a verification result from string.
func ParseVerificationResult(s string) (VerificationResult, error) {
	switch strings.ToLower(s) {
	case string(VerificationResultSuccess):
		return VerificationResultSuccess, nil
	case string(VerificationResultFailed):
		return VerificationResultFailed, nil
	default:
		return "", fmt.Errorf("invalid verification result")
	}
}

type VerificationClaim string

const (
	VerificationClaimEMail   VerificationClaim = "email"
	VerificationClaimName    VerificationClaim = "name"
	VerificationClaimAddress VerificationClaim = "address"
)

// ParseVerificationClaim parses a verification claim from string.
func ParseVerificationClaim(s string) (VerificationClaim, error) {
	switch strings.ToLower(s) {
	case string(VerificationClaimEMail):
		return VerificationClaimEMail, nil
	case string(VerificationClaimName):
		return VerificationClaimName, nil
	case string(VerificationClaimAddress):
		return VerificationClaimAddress, nil
	default:
		return "", fmt.Errorf("invalid verification claim")
	}
}

type VerificationMethod string

const (
	VerificationMethodAuth         VerificationMethod = "auth"
	VerificationMethodEDoc         VerificationMethod = "electronic_document"
	VerificationMethodDoc          VerificationMethod = "physical_document"
	VerificationMethodVDig         VerificationMethod = "vdig"
	VerificationMethodBvr          VerificationMethod = "bvr"
	VerificationMethodPvr          VerificationMethod = "pvr"
	VerificationMethodData         VerificationMethod = "data"
	VerificationMethodReachability VerificationMethod = "reachability"
)

// ParseVerificationMethod parses a verification method from string.
func ParseVerificationMethod(s string) (VerificationMethod, error) {
	switch strings.ToLower(s) {
	case string(VerificationMethodAuth):
		return VerificationMethodAuth, nil
	case string(VerificationMethodEDoc):
		return VerificationMethodEDoc, nil
	case string(VerificationMethodDoc):
		return VerificationMethodDoc, nil
	case string(VerificationMethodVDig):
		return VerificationMethodVDig, nil
	case string(VerificationMethodBvr):
		return VerificationMethodBvr, nil
	case string(VerificationMethodPvr):
		return VerificationMethodPvr, nil
	case string(VerificationMethodData):
		return VerificationMethodData, nil
	case string(VerificationMethodReachability):
		return VerificationMethodReachability, nil
	default:
		return "", fmt.Errorf("invalid verification method")
	}
}

type VerificationEvidence string

const (
	VerificationEvidenceIDCard                  VerificationEvidence = "idcard"
	VerificationEvidencePassport                VerificationEvidence = "passport"
	VerificationEvidencePopulationRegister      VerificationEvidence = "population_register"
	VerificationEvidenceResidencePermit         VerificationEvidence = "residence_permit"
	VerificationEvidenceProofOfArrival          VerificationEvidence = "proof_of_arrival"
	VerificationEvidenceDriversLicence          VerificationEvidence = "drivers_licence"
	VerificationEvidenceCompanyRegister         VerificationEvidence = "company_register"
	VerificationEvidenceCompanyStatement        VerificationEvidence = "company_statement"
	VerificationEvidenceBankAccount             VerificationEvidence = "bank_account"
	VerificationEvidenceOnlinePaymentAccount    VerificationEvidence = "online_payment_account"
	VerificationEvidenceUtilityAccount          VerificationEvidence = "utility_account"
	VerificationEvidenceBankStatement           VerificationEvidence = "bank_statement"
	VerificationEvidenceTaxStatement            VerificationEvidence = "tax_statement"
	VerificationEvidenceWrittenAttestation      VerificationEvidence = "written_attestation"
	VerificationEvidenceDigitalAttestation      VerificationEvidence = "digital_attestation"
	VerificationEvidencePostalVerTransactionLog VerificationEvidence = "postal_ver_transaction_log"
	VerificationEvidenceEmailVerTransactionLog  VerificationEvidence = "email_ver_transaction_log"
	VerificationEvidenceAddressDatabase         VerificationEvidence = "address_database"
)

// ParseVerificationEvidence parses a verification evidence from string.
func ParseVerificationEvidence(s string) (VerificationEvidence, error) {
	switch strings.ToLower(s) {
	case string(VerificationEvidenceIDCard):
		return VerificationEvidenceIDCard, nil
	case string(VerificationEvidencePassport):
		return VerificationEvidencePassport, nil
	case string(VerificationEvidencePopulationRegister):
		return VerificationEvidencePopulationRegister, nil
	case string(VerificationEvidenceResidencePermit):
		return VerificationEvidenceResidencePermit, nil
	case string(VerificationEvidenceProofOfArrival):
		return VerificationEvidenceProofOfArrival, nil
	case string(VerificationEvidenceDriversLicence):
		return VerificationEvidenceDriversLicence, nil
	case string(VerificationEvidenceCompanyRegister):
		return VerificationEvidenceCompanyRegister, nil
	case string(VerificationEvidenceCompanyStatement):
		return VerificationEvidenceCompanyStatement, nil
	case string(VerificationEvidenceBankAccount):
		return VerificationEvidenceBankAccount, nil
	case string(VerificationEvidenceOnlinePaymentAccount):
		return VerificationEvidenceOnlinePaymentAccount, nil
	case string(VerificationEvidenceUtilityAccount):
		return VerificationEvidenceUtilityAccount, nil
	case string(VerificationEvidenceBankStatement):
		return VerificationEvidenceBankStatement, nil
	case string(VerificationEvidenceTaxStatement):
		return VerificationEvidenceTaxStatement, nil
	case string(VerificationEvidenceWrittenAttestation):
		return VerificationEvidenceWrittenAttestation, nil
	case string(VerificationEvidenceDigitalAttestation):
		return VerificationEvidenceDigitalAttestation, nil
	case string(VerificationEvidencePostalVerTransactionLog):
		return VerificationEvidencePostalVerTransactionLog, nil
	case string(VerificationEvidenceEmailVerTransactionLog):
		return VerificationEvidenceEmailVerTransactionLog, nil
	case string(VerificationEvidenceAddressDatabase):
		return VerificationEvidenceAddressDatabase, nil
	default:
		return "", fmt.Errorf("invalid verification evidence")
	}
}

type TrustFramework string

const (
	TrustFrameworkDenic TrustFramework = "de_denic"
)

// ParseTrustFramework parses a trust framework from string.
func ParseTrustFramework(s string) (TrustFramework, error) {
	switch strings.ToLower(s) {
	case string(TrustFrameworkDenic):
		return TrustFrameworkDenic, nil
	default:
		return "", fmt.Errorf("invalid trust framework")
	}
}
