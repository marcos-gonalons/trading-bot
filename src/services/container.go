package services

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler/indicators"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/trends"
)

type Container struct {
	API            api.Interface
	APIRetryFacade retryFacade.Interface
	APIData        api.DataInterface

	Logger                  logger.Interface
	HttpClient              httpclient.Interface
	TrendsService           trends.Interface
	HorizontalLevelsService horizontalLevels.Interface
	IndicatorsService       indicators.MainInterface
}

var container Container

func (c *Container) Initialize() {
	c.Logger = logger.GetInstance()

	c.TrendsService = trends.GetServiceInstance()
	c.HorizontalLevelsService = horizontalLevels.GetServiceInstance()
	c.IndicatorsService = indicators.GetInstance()

	c.HttpClient = httpclient.GetInstance(c.Logger)
}

func (c *Container) SetAPI(API api.Interface) {
	c.API = API

	c.APIRetryFacade = &retryFacade.APIFacade{
		API:    API,
		Logger: c.Logger,
	}

	c.APIData = &api.Data{}
}

func GetServicesContainer() *Container {
	return &container
}
