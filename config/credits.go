package config

import "github.com/anurag-327/neuron/internal/models"

var CreditPricing = map[models.CreditTransactionReason]int64{
	models.CreditReasonSubmission: 5,
	models.CreditReasonRerun:      2,
}

func GetCreditsForReason(reason models.CreditTransactionReason) int64 {
	if v, ok := CreditPricing[reason]; ok {
		return v
	}
	return 0
}
