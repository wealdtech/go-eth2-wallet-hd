// Copyright 2019, 2020 Weald Technology Trading
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package hd_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	types "github.com/wealdtech/go-eth2-wallet-types"
)

func TestCreateWallet(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()
	wallet, err := hd.CreateWallet("test wallet", []byte("wallet passphrase"), store, encryptor)
	assert.Nil(t, err)

	assert.Equal(t, "test wallet", wallet.Name())
	assert.Equal(t, uint(1), wallet.Version())

	// Try to create another wallet with the same name; should fail
	_, err = hd.CreateWallet("test wallet", []byte("wallet passphrase"), store, encryptor)
	assert.NotNil(t, err)

	// Try to obtain the key without unlocking the wallet; should fail
	_, err = wallet.(types.WalletKeyProvider).Key()
	assert.NotNil(t, err)

	err = wallet.Unlock([]byte("wallet passphrase"))
	require.Nil(t, err)

	_, err = wallet.(types.WalletKeyProvider).Key()
	assert.Nil(t, err)
}
