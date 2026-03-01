package basic

// @Channel /market/{pair}
// @ChannelDescription Real-time market data stream for a trading pair
// @ChannelParam pair string true "Trading pair (e.g. BTC-USD)"
// @ChannelServer production
// @ChannelServer staging
// @WsBinding.Method GET
// @WsBinding.Query token string true "Authentication token"
// @WsBinding.Query depth string false "Order book depth" enum(5,10,25)
// @Operation receive
// @OperationID receiveMarketUpdate
// @Summary Receive market data updates
// @Description Streams real-time ticker and order book updates for the subscribed pair
// @Tags market,realtime
// @Message tickerUpdate TickerPayload
// @Message orderBookSnapshot OrderBookPayload
func HandleMarketData() {}

// @Channel /market/{pair}
// @Operation send
// @OperationID sendSubscription
// @Summary Send subscription commands
// @Description Subscribe or unsubscribe from market data channels
// @Tags market
// @Message subscribe SubscribeRequest
// @Reply
// @ReplyMessage subscriptionAck SubscriptionAck
// @ReplyChannel /market/{pair}
func HandleSubscription() {}
