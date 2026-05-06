# Changelog

## [1.0.0] - 2026-05-07

First stable release. Targets AsyncAPI 3.1 (JSON Schema Draft 07). Output is validated against `@asyncapi/parser-js`.

### Changed

- **BREAKING (schema output):** Pointer fields no longer emit `nullable: true` (which is not in JSON Schema Draft 07); nullability is now expressed as `oneOf: [<inner>, {type: "null"}]`. Consumers parsing generated schemas must adapt.
- **BREAKING (schema output):** `[]byte` now maps to `{type: string, format: byte}` (matches `encoding/json` base64), instead of array of integers.
- **BREAKING (schema output):** `interface{}` / `any` now produce `{}` (accept anything) instead of `{type: object}`.
- **BREAKING (Go API):** `spec.Schema.Nullable` field removed.
- **BREAKING (CLI):** Unknown `--outputTypes` value (e.g. `--outputTypes jsn`) now returns an explicit error instead of silently producing nothing.
- **BREAKING (parser):** `findMainFile` no longer falls back to an arbitrary first file when the requested main file isn't found.
- Default AsyncAPI version bumped from `3.0.0` to `3.1.0`.

### Added

- `@SecurityScheme name type [args...] [description]` annotation — `@Security` refs now resolve. Handles `http`, `apiKey`, `httpApiKey`, `openIdConnect`, plus generic types.
- Go 1.18 generics support — `Foo[T]` instantiations resolved by erasing type arguments; type-parameters within generic structs resolve to `{}`.
- `json:",string"` tag option — numeric/bool fields with `,string` are coerced to `{type: string}` to match `encoding/json` wire format.
- `--outputTypes go` automatically includes `json` so `//go:embed asyncapi.json` compiles.
- Channel key collision detection (`ErrChannelKeyCollision`) — `/users/{id}` and `/users/id` no longer silently overwrite each other.
- Duplicate `@Message` detection (`ErrDuplicateMessage`) — same name with different payload across channels now errors.
- Channel data from subsequent handlers (`WsBinding`, `ChannelServers`, `ChannelParams`) is merged into the existing channel; `WsBinding` conflicts produce a warning.
- Tokenizer supports `\"` and `\\` escapes inside quoted annotation arguments.
- `enum("a","b")` / `example("a","b")` strip surrounding quotes; unclosed clauses emit warnings.
- `@WsBinding.Method` warns on values other than `GET`/`POST` (per AsyncAPI WebSocket binding spec).
- Operation message refs (`#/channels/X/messages/Y`) and `op.Reply.Channel` / `op.Reply.Messages` are now validated.
- `gen.Build` aggregates all validation errors via `errors.Join` instead of returning only the first.
- Description on `$ref`-typed fields is preserved by wrapping in `allOf` + `description`.
- Field tags (`min`, `max`, `example`, `description`, etc.) on nullable pointer fields apply to the inner schema, not the `oneOf` wrapper.
- `asyngo.Generate(nil)` returns `ErrNilConfig` instead of panicking.
- CLI version reads from `runtime/debug.ReadBuildInfo` (with `dev` fallback) instead of a hardcoded constant.

### Fixed

- `discriminator` correctly serialized as a string (per AsyncAPI Schema Object spec).
- Pointer to named struct (`*User`) is now nullable (`oneOf: [{$ref}, {type: "null"}]`); previously the `$ref` form lost nullability entirely.
- Unexported struct fields without a json-tag are no longer added to the schema (matches `encoding/json`).
- Embedded `*pkg.Base` correctly flattens its fields into the parent.
- Type catalog is keyed by file directory rather than short package name — `models` packages in different paths no longer collide.
- Deterministic iteration order in handler parsing and type lookup fallback.
- Channel parameters no longer carry a bogus `$message.payload#/...` location.
- Annotation continuation stops on Go-mechanical directives (`go:`, `+build`, `nolint`).
- Embedded external interfaces (`io.Reader` etc.) skip cleanly instead of failing the whole struct.
- Go package name derived from `OutputDir` is sanitized (`my-docs` → `my_docs`, `2docs` → `_2docs`) instead of producing cryptic `go/format` errors.
- `escapeGoString` replaced with `strconv.Quote` so backslashes / special characters in title/version no longer break the generated `docs.go`.

## [0.5.0] - 2026-03-02

### Added

- `oneOf`/`discriminator` support for polymorphic messages via `@MessageOneOf` annotation
- Syntax: `@MessageOneOf <name> <Type1>|<Type2>[|...] [discriminator(<propertyName>)]`
- `Discriminator` field in `spec.Schema` (string, per AsyncAPI 3.0.0 spec)
- Validation of `$ref` entries inside `payload.oneOf`
- `testdata/oneof/` example project with polymorphic event stream
- GitHub Actions CI workflow (`.github/workflows/ci.yml`)
- GitLab CI configuration (`.gitlab-ci.yml`)

### Fixed

- CLI version bumped from `0.2.0` to `0.5.0`

## [0.4.0] - 2026-03-02

### Changed

- **BREAKING:** `Config.SearchDir string` → `Config.SearchDirs []string`, `Config.Excludes string` → `Config.Excludes []string` — comma-separated strings replaced with proper slices
- **BREAKING:** CLI flags `--dir` and `--exclude` now accept multiple values via `StringSliceVar` instead of comma-separated strings
- Extracted `wsBinding` struct from `operationBuilder` — WS binding fields (`Method`, `QueryProps`, `QueryRequired`, `HeaderProps`) are now grouped in a dedicated struct, lazily initialized only when WS annotations are present
- `findMainFile` is now deterministic — `searchDirs` exact match has highest priority, basename match and fallback use sorted candidates instead of random map iteration
- Annotation comment parsing uses `TrimLeft` instead of `TrimPrefix` for leading spaces — correctly handles `//  @Title` (multiple spaces after `//`)
- `ValidateBasic()` (title, version, asyncapi version) now runs unconditionally; full `Validate()` (refs, channels, operations) only runs with `--strict`

### Added

- `spec.AsyncAPI.ValidateBasic()` method for lightweight required-field validation
- Tests for `ValidateBasic`: valid doc, missing individual fields, empty doc, inclusion in full `Validate`

## [0.3.0] - 2026-03-01

### Fixed

- Fixed slice mutation bug in `registerMessages` where `append(b.Messages, b.ReplyMessages...)` could corrupt the original `Messages` slice
- Fixed type name collision in `packagesDefinitions.CatalogTypes` — types from different packages with the same short name no longer overwrite each other
- Fixed `@ChannelParam` silently ignoring `type` argument — unknown types now produce warnings, and `location` field is set
- Fixed potential invalid Go code generation when `Title` or `Version` contain double quotes in `writeGo`
- Fixed `filepath.Match` error being silently ignored in directory exclusion logic
- Fixed `TestBuildUnknownOutputType` not actually verifying unknown format was skipped
- Fixed `testPackageLookup.FindTypeSpec` returning bare `ErrUnresolvedType` instead of wrapping with type name

### Changed

- Removed duplicate code in `resolveArray` — both branches (slice and fixed-length array) were identical
- Refactored `parseHostProtocol` from 7 repetitive if-blocks to table-driven approach using `strings.CutPrefix`
- Removed `funcDeclInfo` wrapper — `*ast.FuncDecl` is used directly
- Removed dead code `isPrimitive` function and its test
- Unknown handler annotations now produce warnings instead of being silently ignored

### Added

- Fuzz tests for `tokenizeArgs` and `parseAnnotationLine`
- Benchmarks for `Parse()`, `tokenizeArgs`, `parseAnnotationLine`, and `ResolveTypeName`
- Edge case tests for tokenizer: unclosed quotes, unclosed parens, empty quotes, parens inside quotes, whitespace-only input
- `t.Parallel()` added to all independent table-driven subtests

## [0.2.0] - 2026-03-01

### Fixed

- Fixed potential panic in `parseChannelParam` when annotation has exactly 3 arguments (`ann.Args[4:]` on a 3-element slice)
- Fixed same panic in `parseWsQueryParam`

### Added

- `mapSimpleType` now reports unknown types instead of silently falling back to `string`
- Non-strict mode: unknown types produce warnings accessible via `Parser.Warnings()`
- Strict mode (`WithStrict(true)`): unknown types return `ErrUnknownType` error
- `Parser.Warnings()` method to retrieve collected warnings after parsing

## [0.1.0] - 2025-12-01

### Added

- Initial release
- AsyncAPI 3.0.0 spec generation from Go annotations
- Channel, operation, message, and schema parsing
- WebSocket binding support (method, query, headers)
- Multi-package project support
- YAML and JSON output
- CLI with `init` and `version` commands
