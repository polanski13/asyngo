package multipackage

// @Channel /events/{userId}
// @ChannelDescription Real-time events for a user
// @ChannelParam userId string true "User identifier"
// @ChannelServer production
// @Operation receive
// @OperationID receiveEvents
// @Summary Receive user events
// @Tags events,realtime
// @Message notification models.Notification
func HandleEvents() {}

// @Channel /events/{userId}
// @Operation send
// @OperationID sendAck
// @Summary Acknowledge events
// @Tags events
// @Message errorResponse models.ErrorResponse
func HandleAck() {}
