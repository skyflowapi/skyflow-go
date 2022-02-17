# Changelog

All notable changes to this project will be documented in this file.

## [1.1.0] - 2022-02-15

### Added
-  Logging functionality
- `SetLogLevel` function for setting the package-level LogLevel
- `GenerateBearerTokenFromCreds` function which takes credentials as string
- `Insert` vault API
- `Detokenize` vault API
- `GetById` vault API
- `InvokeConnection`

### Changed
- Renamed and deprecated `GenerateToken` in favor of `GenerateBearerToken`

## [1.0.0] - 2021-08-25

### Added
-  `GenerateToken` for Service Account Token generation 