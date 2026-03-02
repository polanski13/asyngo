# Changelog

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
