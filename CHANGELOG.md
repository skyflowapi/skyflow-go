# Changelog

All notable changes to this project will be documented in this file.

## [1.3.0] - 2022-03-15

### Changed
- deprecated `IsValid` in favor of `IsExpired`

## [1.2.0] - 2021-02-24

### Added
- Request ID in error logs and error responses for API Errors
- `isValid` method for validating Service Account bearer token

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