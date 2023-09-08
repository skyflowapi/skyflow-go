# Changelog

All notable changes to this project will be documented in this file.

## [1.8.1] - 2023-09-08
### Added
- Added request index in response in Insert Method.

## [1.8.0] - 2023-09-01
### Added
- Support for Bulk request with Continue on Error in Detokenize Method
- Support for Continue on Error in Insert Method

## [1.7.2] - 2023-08-28
### Added
-  Support for OFF Loglevel.

## [1.7.1] - 2023-08-22
### Changed
-  Internal Batch API with tokenization

## [1.7.0] - 2023-08-18
### Added
- Support for BYOT tokens in insert method
- Support for Context in insert method

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
