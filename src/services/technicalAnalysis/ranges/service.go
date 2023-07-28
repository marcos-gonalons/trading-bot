package ranges

import "TradingBot/src/services/technicalAnalysis/horizontalLevels"

type Service struct{}

func (s *Service) GetRange(params GetRangeParams) ([]*horizontalLevels.Level, error) {
	return nil, nil
}

func GetServiceInstance() Interface {
	return &Service{}
}
