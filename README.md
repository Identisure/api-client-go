<img align="left" height="100" style="margin-right: 10px;" src="../img/icon.png">

# IDkollen Go Client

[![pkg.go.dev][pkg-image]][pkg-url]
[![MIT license][license-image]][license-url]

[pkg-image]: https://pkg.go.dev/badge/github.com/idkollen/idk-clients/go/idkollen.svg
[pkg-url]: https://pkg.go.dev/github.com/idkollen/idk-clients/go/idkollen
[license-image]: https://img.shields.io/badge/license-MIT-blue?style=flat-square
[license-url]: https://opensource.org/license/mit

API client for the [IDkollen](https://developers.idkollen.se) REST API.

## Installation

```sh
go get github.com/idkollen/idk-clients/go/idkollen
```

## Usage

```go
package main

import (
    "context"
    "fmt"
    "time"

    idkollen "github.com/idkollen/idk-clients/go/idkollen"
)

func main() {
    client := idkollen.NewClientBuilder("client_id", "client_secret").
        Environment(idkollen.Staging).
        Build()

    ctx := context.Background()

    session, err := client.BankIdSe().Auth(ctx, idkollen.BankIdSeAuthRequest{})
    if err != nil {
        panic(err)
    }

    pending := session.(*idkollen.BankIdSePending)
    fmt.Println("Session ID:", pending.Id)

    pollCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
    defer cancel()

    result, err := client.BankIdSe().WaitForAuth(pollCtx, pending.Id, idkollen.PollOptions{})
    if err != nil {
        panic(err)
    }

    switch r := result.(type) {
    case *idkollen.BankIdSeCompleted:
        fmt.Println("Authenticated:", r.Name, r.Ssn)
    case *idkollen.BankIdSeFailed:
        fmt.Println("Failed:", r.Error)
    }
}
```

## Development

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
