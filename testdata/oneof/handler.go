package oneof

// @Channel /events/{symbol}
// @ChannelDescription Polymorphic event stream for a symbol
// @ChannelParam symbol string true "Trading symbol"
// @Operation receive
// @OperationID receiveEvents
// @Summary Receive market events
// @Tags events
// @MessageOneOf eventUpdate TickerPayload|OrderBookPayload|TradePayload discriminator(eventType)
func HandleEvents() {}
