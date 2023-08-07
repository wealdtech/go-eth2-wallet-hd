// Copyright © 2023 Weald Technology Trading.
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
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestBatch(t *testing.T) {
	ctx := context.Background()

	seed := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	}

	store := scratch.New()
	encryptor := keystorev4.New()

	// Create a wallet.
	wallet, err := hd.CreateWallet(ctx, "test wallet", []byte("wallet passphrase"), store, encryptor, seed)
	require.NoError(t, err)
	require.NoError(t, wallet.(e2wtypes.WalletLocker).Unlock(ctx, []byte("wallet passphrase")))

	// Add some accounts.
	account1, err := wallet.(e2wtypes.WalletAccountCreator).CreateAccount(ctx, "account 1", []byte("passphrase"))
	require.NoError(t, err)
	account2, err := wallet.(e2wtypes.WalletAccountCreator).CreateAccount(ctx, "account 2", []byte("passphrase"))
	require.NoError(t, err)

	// Create a batch.
	require.NoError(t, wallet.(e2wtypes.WalletBatchCreator).BatchWallet(ctx, []string{"passphrase"}, "batch passphrase"))

	// Re-open the wallet and fetch the accounts through the batch system.
	wallet, err = hd.OpenWallet(ctx, "test wallet", store, encryptor)
	require.NoError(t, err)
	numAccounts := 0
	for range wallet.Accounts(ctx) {
		numAccounts++
	}
	require.Equal(t, 2, numAccounts)
	obtainedAccount1, err := wallet.(e2wtypes.WalletAccountByNameProvider).AccountByName(ctx, "account 1")
	require.NoError(t, err)
	require.Equal(t, account1.ID(), obtainedAccount1.ID())
	obtainedAccount2, err := wallet.(e2wtypes.WalletAccountByIDProvider).AccountByID(ctx, account2.ID())
	require.NoError(t, err)
	require.Equal(t, account2.Name(), obtainedAccount2.Name())

	// Ensure we can unlock accounts with the batch passphrase.
	require.NoError(t, obtainedAccount1.(e2wtypes.AccountLocker).Unlock(ctx, []byte("batch passphrase")))
	require.NoError(t, obtainedAccount2.(e2wtypes.AccountLocker).Unlock(ctx, []byte("batch passphrase")))

	// Create another account, not in the batch.
	require.NoError(t, wallet.(e2wtypes.WalletLocker).Unlock(ctx, []byte("wallet passphrase")))
	account3, err := wallet.(e2wtypes.WalletAccountCreator).CreateAccount(ctx, "account 3", []byte("passphrase"))
	require.NoError(t, err)

	// Re-open the wallet and fetch the non-batch account by name.
	wallet, err = hd.OpenWallet(ctx, "test wallet", store, encryptor)
	require.NoError(t, err)
	numAccounts = 0
	for range wallet.Accounts(ctx) {
		numAccounts++
	}
	require.Equal(t, 2, numAccounts)
	obtainedAccount3, err := wallet.(e2wtypes.WalletAccountByNameProvider).AccountByName(ctx, "account 3")
	require.NoError(t, err)
	require.Equal(t, account3.ID(), obtainedAccount3.ID())

	// Re-open the wallet and fetch the non-batch account by ID.
	wallet, err = hd.OpenWallet(ctx, "test wallet", store, encryptor)
	require.NoError(t, err)
	numAccounts = 0
	for range wallet.Accounts(ctx) {
		numAccounts++
	}
	require.Equal(t, 2, numAccounts)
	obtainedAccount3, err = wallet.(e2wtypes.WalletAccountByIDProvider).AccountByID(ctx, account3.ID())
	require.NoError(t, err)
	require.Equal(t, account3.Name(), obtainedAccount3.Name())

	// Ensure we can unlock account with the account passphrase.
	require.NoError(t, obtainedAccount3.(e2wtypes.AccountLocker).Unlock(ctx, []byte("passphrase")))

	// Recreate the batch.
	require.NoError(t, wallet.(e2wtypes.WalletBatchCreator).BatchWallet(ctx, []string{"passphrase", "batch passphrase"}, "batch passphrase"))

	// Re-open the wallet and fetch the accounts through the batch system.
	wallet, err = hd.OpenWallet(ctx, "test wallet", store, encryptor)
	require.NoError(t, err)
	numAccounts = 0
	for range wallet.Accounts(ctx) {
		numAccounts++
	}
	require.Equal(t, 3, numAccounts)
}
