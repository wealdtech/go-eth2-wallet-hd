# go-eth2-wallet-hd

[![Tag](https://img.shields.io/github/tag/wealdtech/go-eth2-wallet-hd.svg)](https://github.com/wealdtech/go-eth2-wallet-hd/releases/)
[![License](https://img.shields.io/github/license/wealdtech/go-eth2-wallet-hd.svg)](LICENSE)
[![GoDoc](https://godoc.org/github.com/wealdtech/go-eth2-wallet-hd?status.svg)](https://godoc.org/github.com/wealdtech/go-eth2-wallet-hd)
[![Travis CI](https://img.shields.io/travis/wealdtech/go-eth2-wallet-hd.svg)](https://travis-ci.org/wealdtech/go-eth2-wallet-hd)
[![codecov.io](https://img.shields.io/codecov/c/github/wealdtech/go-eth2-wallet-hd.svg)](https://codecov.io/github/wealdtech/go-eth2-wallet-hd)
[![Go Report Card](https://goreportcard.com/badge/github.com/wealdtech/go-eth2-wallet-hd)](https://goreportcard.com/report/github.com/wealdtech/go-eth2-wallet-hd)

Hierarchical deterministic [Ethereum 2 wallet](https://github.com/wealdtech/go-eth2-wallet).


## Table of Contents

- [Install](#install)
- [Usage](#usage)
- [Maintainers](#maintainers)
- [Contribute](#contribute)
- [License](#license)

## Install

`go-eth2-wallet-hd` is a standard Go module which can be installed with:

```sh
go get github.com/wealdtech/go-eth2-wallet-hd
```

## Usage

Access to the `wallet` is usually via [go-eth2-wallet](https://github.com/wealdtech/go-eth2-wallet); the first two examples below shows how this can be achieved.

This wallet generates keys deterministically based on a seed phrase stored within the wallet itself.

The wallet uses the path `m/12381/3600/n/0`, where _n_ is the number of the account created; for the first account created _n_ is 0, for the second accout _n_ is 1, _etc._

Wallet and account names may be composed of any valid UTF-8 characters; the only restriction is they can not start with the underscore (`_`) character.

### Example

#### Creating a wallet
```go
package main

import (
	e2wallet "github.com/wealdtech/go-eth2-wallet"
)

func main() {

    // Create a wallet
    wallet, err := e2wallet.CreateWallet("My wallet", e2wallet.WithType("hierarchical deterministic"), e2wallet.WithPassphrase([]byte("my wallet secret")))
    if err != nil {
        panic(err)
    }

    ...
}
```

#### Accessing a wallet
```go
package main

import (
	e2wallet "github.com/wealdtech/go-eth2-wallet"
)

func main() {

    // Open a wallet
    wallet, err := e2wallet.OpenWallet("My wallet")
    if err != nil {
        panic(err)
    }

    ...
}
```

#### Creating an account
```go
package main

import (
	e2wallet "github.com/wealdtech/go-eth2-wallet"
)

func main() {

    // Open a wallet
    wallet, err := e2wallet.OpenWallet("My wallet")
    if err != nil {
        panic(err)
    }

    err = wallet.Unlock([]byte("my wallet secret"))
    if err != nil {
        panic(err)
    }
    // Always immediately defer locking the wallet to ensure it does not remain unlocked outside of the function.
    defer wallet.Lock()
    
    account, err := wallet.CreateAccount("My account", []byte("my account secret"))
    if err != nil {
        panic(err)
    }
    // Wallet should be locked as soon as unlocked operations have finished; it is safe to explicitly call wallet.Lock() as well
    // as defer it as per above.
    wallet.Lock()


    ...
}
```

## Maintainers

Jim McDonald: [@mcdee](https://github.com/mcdee).

## Contribute

Contributions welcome. Please check out [the issues](https://github.com/wealdtech/go-eth2-wallet-hd/issues).

## License

[Apache-2.0](LICENSE) © 2019 Weald Technology Trading Ltd
