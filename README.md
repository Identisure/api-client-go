<img align="left" height="100" style="margin-right: 10px;" src="img/icon.png">

# Identisure Go Client

[![pkg.go.dev][pkg-image]][pkg-url]
[![MIT license][license-image]][license-url]

[pkg-image]: https://pkg.go.dev/badge/github.com/identisure/api-client-go.svg
[pkg-url]: https://pkg.go.dev/github.com/identisure/api-client-go
[license-image]: https://img.shields.io/badge/license-MIT-blue?style=flat-square
[license-url]: https://opensource.org/license/mit

API client for the [Identisure](https://developers.identisure.com) REST API.

## Requirements

Go 1.21 or later.

## Supported providers

- **BankID SE**: auth, sign, phone auth and sign, QR verify, age verification
- **BankID NO**: auth, backchannel auth, sign, age verification
- **Freja eID**: auth, backchannel auth, sign, backchannel sign, age verification
- **MitID**: auth, backchannel auth, sign, age verification
- **FTN** (Finnish Trust Network): auth, age verification
- **Vipps MobilePay**: auth, backchannel auth
- **Documents**: upload and download documents for signing

## Installation

```sh
go get github.com/identisure/api-client-go
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    identisure "github.com/identisure/api-client-go"
)

func main() {
    client := identisure.NewClientBuilder("client_id", "client_secret").
        Environment(identisure.Staging).
        Build()

    ctx := context.Background()

    session, err := client.BankIdSe().Auth(ctx, identisure.BankIdSeAuthRequest{})
    if err != nil {
        panic(err)
    }

    pending := session.(*identisure.BankIdSePending)
    fmt.Println("Session ID:", pending.Id)

    pollCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    result, err := client.BankIdSe().WaitForAuth(pollCtx, pending.Id, identisure.PollOptions{})
    if err != nil {
        panic(err)
    }

    switch r := result.(type) {
    case *identisure.BankIdSeCompleted:
        fmt.Println("Authenticated:", r.Name, r.Ssn)
    case *identisure.BankIdSeFailed:
        fmt.Println("Failed:", r.Error)
    }
}
```

### Verifying callbacks

```go
valid := identisure.VerifyCallbackSignature(rawBody, r.Header.Get("X-Signature-SHA256"), hmacSecret)
```

## Development

Test fixtures live in a submodule. Clone with `--recurse-submodules`, or run
`git submodule update --init` in an existing clone.

#### Prerequisites

- [Go 1.21+](https://go.dev/)

#### Build

```sh
go build ./...
```

#### Test

```sh
go test ./...
```

#### Vet

```sh
go vet ./...
```

## Other languages

[C#](https://github.com/identisure/api-client-csharp), [Dart](https://github.com/identisure/api-client-dart), [TypeScript/JavaScript](https://github.com/identisure/api-client-js), [Java/Kotlin](https://github.com/identisure/api-client-jvm), [PHP](https://github.com/identisure/api-client-php), [Python](https://github.com/identisure/api-client-py), [Ruby](https://github.com/identisure/api-client-ruby), [Rust](https://github.com/identisure/api-client-rust), [Swift](https://github.com/identisure/api-client-swift)
