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
	wallet, err := hd.CreateWallet(context.Background(), "test wallet", []byte("wallet passphrase"), store, encryptor)
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
	wallet, err := hd.CreateWalletFromSeed(context.Background(), "test wallet", []byte("wallet passphrase"), store, encryptor, []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	})
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
			0xb7, 0x36, 0x4b, 0x2c, 0xd1, 0x51, 0x51, 0xde, 0x56, 0xdf, 0x62, 0x78, 0x3c, 0xf8, 0x28, 0xfd,
			0xb7, 0xfe, 0x83, 0xbb, 0xc5, 0x02, 0x46, 0x99, 0xf8, 0x0c, 0x57, 0xa3, 0xa5, 0x32, 0x65, 0xc8,
			0x7c, 0x4c, 0x29, 0x63, 0xae, 0x65, 0xa9, 0x30, 0xd2, 0xed, 0x00, 0xdb, 0xb5, 0xf9, 0x67, 0x82,
		},
		account.PublicKey().Marshal(),
	)
	account, err = accountByNameProvider.AccountByName(context.Background(), "m/12381/3600/1/1/1")
	require.NoError(t, err)
	assert.Equal(t,
		[]byte{
			0xb5, 0x5d, 0x20, 0xfc, 0x82, 0x0f, 0x85, 0x90, 0x6c, 0x88, 0xc0, 0x38, 0x2b, 0xba, 0xad, 0x4e,
			0xf9, 0xe3, 0x66, 0xa5, 0xb4, 0xfc, 0xd1, 0xcf, 0xc2, 0x0f, 0xe0, 0x35, 0xd7, 0x09, 0xfb, 0x56,
			0x1e, 0xea, 0x64, 0xa1, 0xd7, 0xdc, 0xf9, 0xdf, 0x79, 0x38, 0x50, 0xb8, 0x13, 0x4f, 0x1a, 0xe4,
		},
		account.PublicKey().Marshal(),
	)
}
