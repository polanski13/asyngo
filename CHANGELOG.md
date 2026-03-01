# Changelog

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
