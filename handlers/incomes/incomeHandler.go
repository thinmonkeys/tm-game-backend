package incomes

import (
	"../common"
	incomeProvider "../../providers/incomes"
)

func NewHandler(confirmationHandler common.ConfirmationHandler) common.PaymentHandler {
	provider := incomeProvider.NewProvider()
	return common.PaymentHandler {
		ConfirmationHandler: confirmationHandler,
		PaymentLister: provider.GetIncomes,
		PaymentUpdater: provider.SaveIncome,
		Category: common.ScoreCategoryIncomes,
		RequestAuthenticator: common.DefaultRequestAuthenticator().AuthenticateRequestAllowingQueryOverride,
	}
}