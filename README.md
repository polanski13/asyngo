# asyngo

Generate [AsyncAPI 3.0](https://www.asyncapi.com/) specifications from Go source code annotations. Like [swaggo/swag](https://github.com/swaggo/swag) but for WebSocket and event-driven APIs.

## Install

```bash
go install github.com/polanski13/asyngo/cmd/asyngo@latest
```

## Quick Start

1. Annotate your main function with general API info:

```go
// @AsyncAPI 3.0.0
// @Title Trading Platform WebSocket API
// @Version 1.0.0
// @Description Real-time trading data streaming
// @DefaultContentType application/json
// @Server production wss://ws.example.com /v1 "Production endpoint"
func main() {}
```

2. Annotate handler functions with channel and operation info:

```go
// @Channel /market/{pair}
// @ChannelDescription Real-time market data stream
// @ChannelParam pair string true "Trading pair"
// @WsBinding.Method GET
// @WsBinding.Query token string true "Auth token"
// @Operation receive
// @OperationID receiveMarketData
// @Summary Receive market updates
// @Tags market,realtime
// @Message tickerUpdate TickerPayload
// @Message orderBook OrderBookPayload
func HandleMarket(conn *websocket.Conn) {}
```

3. Use `@MessageOneOf` for polymorphic messages with multiple payload types:

```go
// @Channel /events/{symbol}
// @Operation receive
// @OperationID receiveEvents
// @MessageOneOf eventUpdate TickerPayload|OrderBookPayload|TradePayload discriminator(eventType)
func HandleEvents() {}
```

This generates a `oneOf` payload with a `discriminator` field:

```yaml
payload:
  oneOf:
    - $ref: '#/components/schemas/TickerPayload'
    - $ref: '#/components/schemas/OrderBookPayload'
    - $ref: '#/components/schemas/TradePayload'
  discriminator: eventType
```

4. Define payload types as Go structs:

```go
type TickerPayload struct {
    Symbol string    `json:"symbol" validate:"required" example:"BTC-USD"`
    Price  float64   `json:"price" minimum:"0"`
    Side   string    `json:"side" enum:"buy,sell"`
    Time   time.Time `json:"timestamp"`
}
```

5. Generate:

```bash
asyngo init --dir . --output ./docs
```

This creates `asyncapi.json` and `asyncapi.yaml` in `./docs`.

## CLI

```
asyngo init [flags]

Flags:
  -d, --dir string           Search directories (default ".")
      --main string          Go file with general API annotations (default "main.go")
  -o, --output string        Output directory (default "./docs")
      --outputTypes strings  Output types: json, yaml, go (default [json,yaml])
      --exclude string       Exclude directories (comma-separated)
      --strict               Treat warnings as errors, enable validation
```

## Programmatic Usage

```go
import "github.com/polanski13/asyngo"

err := asyngo.Generate(&asyngo.Config{
    SearchDir:   "./",
    MainAPIFile: "main.go",
    OutputDir:   "./docs",
    OutputTypes: []string{"json", "yaml"},
})
```

## Annotation Reference

### General API Info

| Annotation | Description |
|---|---|
| `@AsyncAPI` | AsyncAPI spec version (e.g. `3.0.0`) |
| `@Title` | API title |
| `@Version` | API version |
| `@Description` | API description (supports continuation lines) |
| `@DefaultContentType` | Default content type (e.g. `application/json`) |
| `@ID` | Unique document identifier |
| `@TermsOfService` | Terms of service URL |
| `@Contact.Name` | Contact name |
| `@Contact.Email` | Contact email |
| `@Contact.URL` | Contact URL |
| `@License.Name` | License name |
| `@License.URL` | License URL |
| `@ExternalDocs.Description` | External docs description |
| `@ExternalDocs.URL` | External docs URL |
| `@Server` | Server definition: `name host pathname "description"` |

### Channel & Operation

| Annotation | Description |
|---|---|
| `@Channel` | Channel address (e.g. `/market/{pair}`) |
| `@ChannelDescription` | Channel description |
| `@ChannelParam` | Channel parameter: `name type required "description"` |
| `@ChannelServer` | Associate channel with a server |
| `@WsBinding.Method` | WebSocket HTTP method (GET, POST) |
| `@WsBinding.Query` | WebSocket query parameter: `name type required "description"` |
| `@WsBinding.Header` | WebSocket header: `name type required "description"` |
| `@Operation` | Operation action: `send` or `receive` |
| `@OperationID` | Unique operation identifier |
| `@Summary` | Operation summary |
| `@Description` | Operation description |
| `@Tags` | Comma-separated tags |
| `@Message` | Message definition: `name PayloadType` |
| `@MessageOneOf` | Polymorphic message: `name Type1\|Type2 [discriminator(prop)]` |
| `@Reply` | Mark operation as having a reply |
| `@ReplyMessage` | Reply message: `name PayloadType` |
| `@ReplyChannel` | Reply channel address |
| `@Security` | Security scheme reference |

### Struct Tags

| Tag | Description | Example |
|---|---|---|
| `json` | JSON field name, `-` to skip | `json:"name"` |
| `validate` | Validation rules (`required`, `min=`, `max=`, `oneof=`) | `validate:"required,min=1"` |
| `binding` | Alternative validation (`required`) | `binding:"required"` |
| `example` | Example value | `example:"BTC-USD"` |
| `enum` | Comma-separated enum values | `enum:"buy,sell"` |
| `minimum` | Minimum value | `minimum:"0"` |
| `maximum` | Maximum value | `maximum:"100"` |
| `format` | JSON Schema format | `format:"email"` |
| `pattern` | Regex pattern | `pattern:"^[a-z]+$"` |
| `default` | Default value | `default:"auto"` |
| `asyncapiignore` | Exclude field from schema | `asyncapiignore:"true"` |

## Supported Types

| Go Type | JSON Schema |
|---|---|
| `string` | `string` |
| `int`, `int8`..`int64`, `uint`..`uint64` | `integer` |
| `float32`, `float64` | `number` |
| `bool` | `boolean` |
| `time.Time` | `string` (format: `date-time`) |
| `uuid.UUID` | `string` (format: `uuid`) |
| `[]T` | `array` (items: T) |
| `map[string]T` | `object` (additionalProperties: T) |
| `*T` | T (nullable: true) |
| embedded struct | flattened into parent |
| `any`, `interface{}` | `object` |

## Supported Protocols

`wss://`, `ws://`, `https://`, `http://`, `mqtt://`, `amqp://`, `kafka://`

If no protocol prefix is specified, `ws` is assumed.

## Channel Merging

Multiple handlers can annotate the same `@Channel` address. Messages from all handlers are merged into a single channel definition. The first handler's description and bindings take precedence.

## License

MIT
