package services

import (
	"TradingBot/src/services/api"
	"TradingBot/src/services/api/retryFacade"
	"TradingBot/src/services/candlesHandler/indicators"
	"TradingBot/src/services/httpclient"
	"TradingBot/src/services/logger"
	"TradingBot/src/services/logger/nullLogger"
	"TradingBot/src/services/positionSize"
	"TradingBot/src/services/technicalAnalysis/horizontalLevels"
	"TradingBot/src/services/technicalAnalysis/ranges"
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
	RangesService           ranges.Interface
	IndicatorsService       indicators.MainInterface
	PositionSizeService     positionSize.Interface
}

var container Container

func (c *Container) Initialize(logging bool) {
	if logging {
		c.Logger = logger.GetInstance()
	} else {
		c.Logger = nullLogger.GetInstance()
	}

	c.HttpClient = httpclient.GetInstance(c.Logger)
	c.TrendsService = trends.GetServiceInstance()
	c.HorizontalLevelsService = horizontalLevels.GetServiceInstance()
	c.RangesService = ranges.GetServiceInstance()
	c.IndicatorsService = indicators.GetInstance()
	c.PositionSizeService = positionSize.GetInstance()
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
