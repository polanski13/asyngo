<p align="center">
  <img src="./.github/assets/asyngo-hero.svg" alt="asyngo turns Go source annotations into AsyncAPI 3.1 documents" width="100%">
</p>

<h1 align="center">asyngo</h1>

<p align="center">
  <strong>Generate AsyncAPI 3.1 contracts directly from annotated Go source.</strong>
</p>

<p align="center">
  <a href="https://github.com/polanski13/asyngo/actions/workflows/ci.yml"><img src="https://img.shields.io/github/actions/workflow/status/polanski13/asyngo/ci.yml?branch=main&style=flat-square&label=build" alt="Build status"></a>
  <a href="https://github.com/polanski13/asyngo/releases"><img src="https://img.shields.io/github/v/release/polanski13/asyngo?style=flat-square&sort=semver" alt="Latest release"></a>
  <a href="https://github.com/polanski13/asyngo/blob/main/go.mod"><img src="https://img.shields.io/github/go-mod/go-version/polanski13/asyngo?style=flat-square" alt="Go version"></a>
  <a href="https://www.asyncapi.com/docs/reference/specification/v3.1.0"><img src="https://img.shields.io/badge/AsyncAPI-3.1-8b7cff?style=flat-square" alt="AsyncAPI 3.1"></a>
  <a href="#license"><img src="https://img.shields.io/badge/license-MIT-00add8?style=flat-square" alt="MIT license"></a>
</p>

<p align="center">
  <a href="#quick-start">Quick start</a> ·
  <a href="#annotation-reference">Annotations</a> ·
  <a href="#cli-reference">CLI</a> ·
  <a href="./CHANGELOG.md">Changelog</a>
</p>

---

`asyngo` keeps event-driven API documentation beside the Go code that implements it. Add familiar comment annotations to your API metadata, channels, operations, and payload structs, then generate validated AsyncAPI in YAML, JSON, or embeddable Go.

If [`swaggo/swag`](https://github.com/swaggo/swag) is how you document REST endpoints, `asyngo` brings the same source-first workflow to WebSocket and asynchronous APIs.

## Why asyngo

- **One source of truth.** Channels, messages, bindings, and schemas live beside their Go handlers and types.
- **AsyncAPI 3.1 output.** Generated documents use JSON Schema Draft 07 and are checked in CI with the official AsyncAPI CLI.
- **Event-native modeling.** Describe send and receive operations, replies, WebSocket bindings, polymorphic messages, and security schemes.
- **CI-friendly generation.** Emit deterministic JSON, YAML, or Go with strict validation when warnings should fail the build.

### Where it fits

| Tool | Direction | Best suited for |
|---|---|---|
| **asyngo** | Go annotations → AsyncAPI 3.1 | Keeping WebSocket and event contracts in Go source |
| [`swaggo/swag`](https://github.com/swaggo/swag) | Go annotations → Swagger 2.0 | Documenting REST APIs from Go source |
| [`asyncapi-codegen`](https://github.com/lerenn/asyncapi-codegen) | AsyncAPI document → Go | Generating broker and application code from an existing contract |

Choose `asyngo` when the Go code is authoritative and the AsyncAPI document should be generated from it.

## Quick start

Requires Go 1.25 or later.

### 1. Install

```bash
go install github.com/polanski13/asyngo/cmd/asyngo@latest
```

### 2. Add API metadata

Place the general annotations in your main API file:

```go
// @AsyncAPI 3.1.0
// @Title Trading WebSocket API
// @Version 1.0.0
// @DefaultContentType application/json
// @Server production wss://ws.example.com /v1 "Production"
func main() {}
```

### 3. Describe a channel

Annotate the handler that sends or receives messages:

```go
// @Channel /market/{pair}
// @ChannelParam pair string true "Trading pair"
// @WsBinding.Query token string true "Authentication token"
// @Operation receive
// @OperationID receiveMarketData
// @Summary Receive live market updates
// @Message tickerUpdate TickerPayload
func HandleMarket() {}
```

Define the message payload as a regular Go struct:

```go
type TickerPayload struct {
    Symbol string    `json:"symbol" validate:"required" example:"BTC-USD"`
    Price  float64   `json:"price" minimum:"0"`
    Side   string    `json:"side" enum:"buy,sell"`
    Time   time.Time `json:"timestamp"`
}
```

### 4. Generate

```bash
asyngo init --dir . --output ./docs --strict
```

By default, `asyngo` writes `asyncapi.json` and `asyncapi.yaml`:

```yaml
asyncapi: 3.1.0
channels:
  marketPair:
    address: /market/{pair}
    messages:
      tickerUpdate:
        $ref: '#/components/messages/tickerUpdate'
operations:
  receiveMarketData:
    action: receive
    channel:
      $ref: '#/channels/marketPair'
```

## Model event-driven APIs

### Polymorphic messages

Use `@MessageOneOf` when a channel carries multiple payload shapes under one message:

```go
// @Channel /events/{symbol}
// @Operation receive
// @OperationID receiveEvents
// @MessageOneOf eventUpdate TickerPayload|OrderBookPayload|TradePayload discriminator(eventType)
func HandleEvents() {}
```

The resulting message payload contains `oneOf` references and an AsyncAPI discriminator:

```yaml
payload:
  oneOf:
    - $ref: '#/components/schemas/TickerPayload'
    - $ref: '#/components/schemas/OrderBookPayload'
    - $ref: '#/components/schemas/TradePayload'
  discriminator: eventType
```

### Request and reply

Mark an operation with `@Reply`, then identify its message and channel:

```go
// @Channel /market/{pair}
// @Operation send
// @OperationID sendSubscription
// @Message subscribe SubscribeRequest
// @Reply
// @ReplyMessage subscriptionAck SubscriptionAck
// @ReplyChannel /market/{pair}
func HandleSubscription() {}
```

### Security schemes

Declare schemes at the API level and reference them from operations:

```go
// @SecurityScheme bearerAuth http bearer JWT "Bearer token"
func main() {}

// @Channel /account/events
// @Operation receive
// @OperationID receiveAccountEvents
// @Security bearerAuth
// @Message accountUpdate AccountPayload
func HandleAccountEvents() {}
```

Supported security scheme types include `http`, `apiKey`, `httpApiKey`, and `openIdConnect`.

## CLI reference

```text
asyngo init [flags]

Flags:
  -d, --dir strings          Directories to search (default [.])
      --main string          Go file with general API annotations (default "main.go")
  -o, --output string        Output directory (default "./docs")
      --outputTypes strings  Output types: json, yaml, go (default [json,yaml])
      --exclude strings      Directories to exclude
      --strict               Treat warnings as errors
```

Specify multiple search or exclude directories as comma-separated values:

```bash
asyngo init \
  --dir ./cmd/server,./internal/realtime \
  --exclude ./internal/generated,./vendor \
  --outputTypes yaml,go \
  --strict
```

When `go` output is selected, `asyngo` also writes `asyncapi.json` so the generated package can embed it.

## Programmatic usage

```go
package docs

import (
    "github.com/polanski13/asyngo"
    "github.com/polanski13/asyngo/gen"
)

func Generate() error {
    return asyngo.Generate(&gen.Config{
        SearchDirs:  []string{"."},
        MainAPIFile: "main.go",
        OutputDir:   "./docs",
        OutputTypes: []string{"json", "yaml"},
        Strict:      true,
    })
}
```

## Annotation reference

### General API

| Annotation | Description |
|---|---|
| `@AsyncAPI` | AsyncAPI version, for example `3.1.0` |
| `@Title` | API title |
| `@Version` | API version |
| `@Description` | API description; continuation lines are supported |
| `@DefaultContentType` | Default content type, for example `application/json` |
| `@ID` | Unique document identifier |
| `@TermsOfService` | Terms of service URL |
| `@Contact.Name` | Contact name |
| `@Contact.Email` | Contact email |
| `@Contact.URL` | Contact URL |
| `@License.Name` | License name |
| `@License.URL` | License URL |
| `@ExternalDocs.Description` | External documentation description |
| `@ExternalDocs.URL` | External documentation URL |
| `@Server` | `name host pathname "description"` |
| `@SecurityScheme` | `name type [type-specific arguments] ["description"]` |

### Channels and operations

| Annotation | Description |
|---|---|
| `@Channel` | Channel address, for example `/market/{pair}` |
| `@ChannelDescription` | Channel description |
| `@ChannelParam` | `name type required "description"` |
| `@ChannelServer` | Associate the channel with a named server |
| `@WsBinding.Method` | WebSocket HTTP method: `GET` or `POST` |
| `@WsBinding.Query` | `name type required "description"` |
| `@WsBinding.Header` | `name type required "description"` |
| `@Operation` | Operation action: `send` or `receive` |
| `@OperationID` | Unique operation identifier |
| `@Summary` | Operation summary |
| `@Description` | Operation description |
| `@Tags` | Comma-separated tags |
| `@Message` | `name PayloadType` |
| `@MessageOneOf` | `name Type1\|Type2 [discriminator(property)]` |
| `@Reply` | Mark the operation as having a reply |
| `@ReplyMessage` | `name PayloadType` |
| `@ReplyChannel` | Reply channel address |
| `@Security` | Security scheme reference |

### Struct tags

| Tag | Description | Example |
|---|---|---|
| `json` | JSON field name; use `-` to skip | `json:"name"` |
| `validate` | Validation rules | `validate:"required,min=1"` |
| `binding` | Alternative required-field marker | `binding:"required"` |
| `example` | Example value | `example:"BTC-USD"` |
| `enum` | Comma-separated enum values | `enum:"buy,sell"` |
| `minimum` | Minimum numeric value | `minimum:"0"` |
| `maximum` | Maximum numeric value | `maximum:"100"` |
| `format` | JSON Schema format | `format:"email"` |
| `pattern` | Regular expression | `pattern:"^[a-z]+$"` |
| `default` | Default value | `default:"auto"` |
| `asyncapiignore` | Exclude the field from its schema | `asyncapiignore:"true"` |

## Type mapping

| Go type | Generated JSON Schema |
|---|---|
| `string` | `string` |
| `int`, `int8`…`int64`, `uint`…`uint64` | `integer` |
| `float32`, `float64` | `number` |
| `bool` | `boolean` |
| `time.Time` | `string`, format `date-time` |
| `uuid.UUID` | `string`, format `uuid` |
| `[]byte` | `string`, format `byte` |
| `[]T` | `array` with `T` items |
| `map[string]T` | `object` with `T` values |
| `*T` | `oneOf` with `T` and `null` |
| Embedded struct | Fields flattened into the parent |
| `any`, `interface{}` | Unconstrained schema |

Generic type instantiations are supported. Type parameters inside a generic struct resolve to an unconstrained schema.

## Protocols

`wss://` · `ws://` · `https://` · `http://` · `mqtt://` · `amqp://` · `kafka://`

When a server host has no protocol prefix, `ws` is used.

## Channel merging

Multiple handlers can annotate the same `@Channel` address. `asyngo` merges their messages, parameters, servers, and WebSocket bindings into one channel definition. Conflicting WebSocket bindings produce a warning, which becomes an error in strict mode.

## Development

```bash
go test ./...
go vet ./...
go build ./...
```

Generated fixtures are also validated in CI with [`@asyncapi/cli`](https://github.com/asyncapi/cli).

## License

MIT
