package mainstrategy

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/ibroker"
	"TradingBot/src/utils"
	"time"
)

func (s *Strategy) closeSpecificOrders(
	orders []*api.Order,
	successCallback func(),
	onErrorCallback func(err error),
) {
	if orders == nil || len(orders) == 0 {
		successCallback()
		return
	}

	s.Logger.Log("Closing specified orders -> " + utils.GetStringRepresentation(orders))

	go utils.RepeatUntilSuccess(
		"CloseSpecifiedOrders",
		func() (err error) {
			for _, order := range orders {
				err = s.API.CloseOrder(order.ID)
				if err != nil {
					s.Logger.Error("An error happened while closing the specified orders -> " + err.Error())
					onErrorCallback(err)
					return
				}
			}
			return
		},
		5*time.Second,
		30,
		successCallback,
	)
}

func (s *Strategy) closeAllWorkingOrders(
	successCallback func(),
	onErrorCallback func(err error),
) {
	s.Logger.Log("Closing all working orders ...")

	go utils.RepeatUntilSuccess(
		"CloseAllWorkingOrders",
		func() (err error) {
			err = s.API.CloseAllOrders()
			if err != nil {
				s.Logger.Error("An error happened while closing all orders -> " + err.Error())
				onErrorCallback(err)
			}
			return
		},
		5*time.Second,
		30,
		successCallback,
	)
}

func (s *Strategy) closePositions(
	successCallback func(),
	onErrorCallback func(err error),
) {
	s.Logger.Log("Closing all positions ...")

	go utils.RepeatUntilSuccess(
		"CloseAllPositions",
		func() (err error) {
			err = s.API.CloseAllPositions()
			if err != nil {
				s.Logger.Error("An error happened while closing all positions -> " + err.Error())
				onErrorCallback(err)
			}
			return
		},
		5*time.Second,
		30,
		successCallback,
	)
}

func (s *Strategy) modifyPosition(tp string, sl string) {
	s.Logger.Log("Modifying the current open position with this values: tp -> " + tp + " and sl -> " + sl)

	go utils.RepeatUntilSuccess(
		"ModifyPosition",
		func() (err error) {
			err = s.API.ModifyPosition(ibroker.GER30SymbolName, &tp, &sl)
			if err != nil {
				s.Logger.Error("An error happened while modifying the position -> " + err.Error())
			}
			return
		},
		5*time.Second,
		20,
		func() {},
	)
}
