package standingorders

import (
	"../common"
	soProvider "../../providers/standingorders"
)

func NewHandler(confirmationHandler common.ConfirmationHandler) common.PaymentHandler {
	provider := soProvider.NewProvider()
	return common.PaymentHandler {
		ConfirmationHandler: confirmationHandler,
		PaymentLister: provider.GetStandingOrders,
		PaymentUpdater: provider.SaveStandingOrder,
		Category: common.ScoreCategoryStandingOrders,
		RequestAuthenticator: common.DefaultRequestAuthenticator().AuthenticateRequestAllowingQueryOverride,
	}
}