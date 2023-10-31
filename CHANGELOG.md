# Changelog

All notable changes to this project will be documented in this file.

## [1.10.0] - 2023-10-31
### Added
- `Get` method 

## [1.6.0] - 2023-06-09
### Added
- `redaction` key for detokenize method for column group support.

## [1.5.1] - 2023-03-01
### Added
- Fix token expiry time and removal of grace period.

## [1.5.0] - 2022-12-07
### Added
- Upsert support for `insert` method.


## [1.4.0] - 2022-04-12

### Added
- Support for application/x-www-form-urlencoded and multipart/form-data content-type's in connections.

## [1.3.1] - 2022-03-29

### Changed
- Added validation to token from TokenProvider

### Fixed 
-  requestHeaders are not case insensitive

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
