package positionSize

type Strategy uint8

const (
	BASED_ON_STOP_LOSS_DISTANCE Strategy = 0
	BASED_ON_MULTIPLIER         Strategy = 1
	BASED_ON_MIN_SIZE           Strategy = 2
)
