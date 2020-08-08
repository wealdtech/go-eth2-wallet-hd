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
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestCreateAccount(t *testing.T) {
	tests := []struct {
		name              string
		accountName       string
		walletPassphrase  []byte
		accountPassphrase []byte
		err               error
	}{
		{
			name:        "Empty",
			accountName: "",
			err:         errors.New("account name missing"),
		},
		{
			name:        "Invalid",
			accountName: "_bad",
			err:         errors.New(`invalid account name "_bad"`),
		},
		{
			name:        "Good",
			accountName: "test",
		},
		{
			name:        "Duplicate",
			accountName: "test",
			err:         errors.New(`account with name "test" already exists`),
		},
	}

	store := scratch.New()
	encryptor := keystorev4.New()
	seed := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	}
	wallet, err := hd.CreateWallet(context.Background(), "test wallet", []byte("wallet passphrase"), store, encryptor, seed)
	require.Nil(t, err)

	// Try to create without unlocking the wallet; should fail.
	_, err = wallet.(e2wtypes.WalletAccountCreator).CreateAccount(context.Background(), "attempt", []byte("test"))
	assert.NotNil(t, err)

	locker, isLocker := wallet.(e2wtypes.WalletLocker)
	require.True(t, isLocker)
	err = locker.Unlock(context.Background(), []byte("wallet passphrase"))
	require.Nil(t, err)
	defer func(t *testing.T) {
		require.NoError(t, locker.Lock(context.Background()))
	}(t)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err = wallet.(e2wtypes.WalletAccountCreator).CreateAccount(context.Background(), test.accountName, test.accountPassphrase)
			if test.err != nil {
				require.NotNil(t, err)
				assert.Equal(t, test.err.Error(), err.Error())
			} else {
				require.Nil(t, err)
				accountByNameProvider, isAccountByNameProvider := wallet.(e2wtypes.WalletAccountByNameProvider)
				require.True(t, isAccountByNameProvider)
				account, err := accountByNameProvider.AccountByName(context.Background(), test.accountName)
				require.Nil(t, err)
				assert.Equal(t, test.accountName, account.Name())
				pathProvider, isPathProvider := account.(e2wtypes.AccountPathProvider)
				require.True(t, isPathProvider)
				assert.NotNil(t, pathProvider.Path())
				require.Equal(t, wallet.Name(), account.(e2wtypes.AccountWalletProvider).Wallet().Name())

				// Should not be able to obtain private key from a locked account.
				_, err = account.(e2wtypes.AccountPrivateKeyProvider).PrivateKey(context.Background())
				assert.NotNil(t, err)
				locker, isLocker := account.(e2wtypes.AccountLocker)
				require.True(t, isLocker)
				err = locker.Unlock(context.Background(), test.accountPassphrase)
				require.Nil(t, err)
				_, err = account.(e2wtypes.AccountPrivateKeyProvider).PrivateKey(context.Background())
				assert.Nil(t, err)
			}
		})
	}
}

func TestAccountByNameDynamic(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()
	seed := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	}
	wallet, err := hd.CreateWallet(context.Background(), "test wallet", []byte("wallet passphrase"), store, encryptor, seed)
	require.NoError(t, err)

	locker, isLocker := wallet.(e2wtypes.WalletLocker)
	require.True(t, isLocker)
	err = locker.Unlock(context.Background(), []byte("wallet passphrase"))
	require.NoError(t, err)

	accountByNameProvider, isAccountByNameProvider := wallet.(e2wtypes.WalletAccountByNameProvider)
	require.True(t, isAccountByNameProvider)
	account, err := accountByNameProvider.AccountByName(context.Background(), "m/12381/3600/0/0")
	require.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0x8a, 0x03, 0xfe, 0x6a, 0x38, 0x53, 0xaa, 0x04, 0xc5, 0xac, 0xaf, 0x31, 0x79, 0xd1, 0xa7, 0xc5,
			0x23, 0x5c, 0xda, 0xaf, 0xbd, 0xc0, 0xdf, 0xec, 0xcd, 0x3e, 0x25, 0xde, 0xf5, 0x38, 0xa5, 0xfc,
			0xe4, 0xaf, 0xb6, 0xb4, 0x9a, 0x7b, 0x3d, 0x45, 0x7d, 0xf1, 0x23, 0x32, 0x6a, 0x3c, 0x13, 0x5f,
		},
		account.PublicKey().Marshal(),
	)
	account, err = accountByNameProvider.AccountByName(context.Background(), "m/12381/3600/1/1/1")
	require.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0x8b, 0x7f, 0x2c, 0x6c, 0x2c, 0x3f, 0xab, 0x1e, 0x14, 0xca, 0xdf, 0x12, 0x44, 0xd3, 0x4e, 0xcf,
			0x94, 0xc6, 0x05, 0xa5, 0x8f, 0x01, 0x3e, 0xb8, 0x3b, 0x08, 0xa3, 0xbb, 0x39, 0xb1, 0x91, 0x74,
			0x98, 0x0d, 0x08, 0xfb, 0xf9, 0x0e, 0x12, 0xab, 0x8d, 0x5d, 0xab, 0x44, 0x15, 0x3d, 0x33, 0x7a,
		},
		account.PublicKey().Marshal(),
	)
}

func TestCreatePathedAccount(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()
	seed := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	}
	wallet, err := hd.CreateWallet(context.Background(), "test wallet", []byte("wallet passphrase"), store, encryptor, seed)
	require.Nil(t, err)
	locker, isLocker := wallet.(e2wtypes.WalletLocker)
	require.True(t, isLocker)
	err = locker.Unlock(context.Background(), []byte("wallet passphrase"))
	require.Nil(t, err)

	// Create an account without a path.
	_, err = wallet.(e2wtypes.WalletAccountCreator).CreateAccount(context.Background(), "Test", []byte("account passphrase"))
	require.Nil(t, err)
	// Attempt to create an account with the same path; should fail.
	_, err = wallet.(e2wtypes.WalletPathedAccountCreator).CreatePathedAccount(context.Background(), "m/12381/3600/0/0", "Test 2", []byte("account passphrase"))
	require.EqualError(t, err, `account with path "m/12381/3600/0/0" already exists`)

	// Attempt to create an account with the a different path; should succeed.
	_, err = wallet.(e2wtypes.WalletPathedAccountCreator).CreatePathedAccount(context.Background(), "m/12381/3600/1/2/3", "Test 3", []byte("account passphrase"))
	require.Nil(t, err)
	// Attempt to create an account with the the same path; should fail.
	_, err = wallet.(e2wtypes.WalletPathedAccountCreator).CreatePathedAccount(context.Background(), "m/12381/3600/1/2/3", "Test 4", []byte("account passphrase"))
	require.EqualError(t, err, `account with path "m/12381/3600/1/2/3" already exists`)
}

func TestCreatePathedAccountConflict(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()
	seed := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	}
	wallet, err := hd.CreateWallet(context.Background(), "test wallet", []byte("wallet passphrase"), store, encryptor, seed)
	require.Nil(t, err)
	locker, isLocker := wallet.(e2wtypes.WalletLocker)
	require.True(t, isLocker)
	err = locker.Unlock(context.Background(), []byte("wallet passphrase"))
	require.Nil(t, err)

	// Create an account with the explicit path of the first index.
	_, err = wallet.(e2wtypes.WalletPathedAccountCreator).CreatePathedAccount(context.Background(), "m/12381/3600/0/0", "Test 1", []byte("account passphrase"))
	require.Nil(t, err)

	// Now create an unpathed account; should have the next index.
	account, err := wallet.(e2wtypes.WalletAccountCreator).CreateAccount(context.Background(), "Test 2", []byte("account passphrase"))
	require.Nil(t, err)
	require.Equal(t, "m/12381/3600/1/0", account.(e2wtypes.AccountPathProvider).Path())

	// Now create another unpathed account; should have the next index.
	account, err = wallet.(e2wtypes.WalletAccountCreator).CreateAccount(context.Background(), "Test 3", []byte("account passphrase"))
	require.Nil(t, err)
	require.Equal(t, "m/12381/3600/2/0", account.(e2wtypes.AccountPathProvider).Path())
}
