package constants

const GER30SymbolName = "GER30"

const SP500SymbolName = "SPX500"

const EUROSTOXXSymbolName = "EUSTX50"

const EURUSDSymbolName = "EUR/USD"

const GBPUSDSymbolName = "GBP/USD"

const USDCADSymbolName = "USD/CAD"

const USDJPYSymbolName = "USD/JPY"

const USDCHFSymbolName = "USD/CHF"

const NZDUSDSymbolName = "NZD/USD"

const AUDUSDSymbolName = "AUD/USD"

var SessionDisconnectedErrorStrings = []string{
	"Your session is disconnected. Please login again to initialize a new valid session.",
	"se encuentra desconectada",
}

const OrderAlreadyExistsErrorString = "ya existe alguna orden vigente"

const PositionAlreadyExistsErrorString = "No es posible introducir una nueva orden con bracket TP/SL"

const NotEnoughFundsErrorString = "Saldo insuficiente"

const OrderIsPendingCancelErrorString = "Pending Cancel"

const OrderIsCancelledErrorString = "Order Status is Cancelled"

const OrderIsFilledErrorString = "Order Status is Filled"

const InvalidHoursErrorString = "Horario incorrecto"

const InvalidHoursErrorString2 = "cerrada a esta hora"

const InvalidHoursErrorString3 = "Contrato en fase Closed"

const ClosePositionRequestInProgressErrorString = "Orden de cerrar no procesada porque existe otra orden en curso"

const PositionNotFoundError = "no tiene posición abierta en el contrato"

const StatusWorkingOrder = "working"

const LongSide = "buy"

const ShortSide = "sell"

const LimitType = "limit"

const StopType = "stop"

const MarketType = "market"
