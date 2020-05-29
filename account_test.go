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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	hd "github.com/wealdtech/go-eth2-wallet-hd/v2"
	scratch "github.com/wealdtech/go-eth2-wallet-store-scratch"
	wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
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
	wallet, err := hd.CreateWallet("test wallet", []byte("wallet passphrase"), store, encryptor)
	require.Nil(t, err)

	// Try to create without unlocking the wallet; should fail.
	_, err = wallet.CreateAccount("attempt", []byte("test"))
	assert.NotNil(t, err)

	err = wallet.Unlock([]byte("wallet passphrase"))
	require.Nil(t, err)
	defer wallet.Lock()
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err = wallet.CreateAccount(test.accountName, test.accountPassphrase)
			if test.err != nil {
				require.NotNil(t, err)
				assert.Equal(t, test.err.Error(), err.Error())
			} else {
				require.Nil(t, err)
				account, err := wallet.AccountByName(test.accountName)
				require.Nil(t, err)
				assert.Equal(t, test.accountName, account.Name())
				assert.NotNil(t, account.Path())

				// Should not be able to obtain private key from a locked account.
				_, err = account.(wtypes.AccountPrivateKeyProvider).PrivateKey()
				assert.NotNil(t, err)
				err = account.Unlock(test.accountPassphrase)
				require.Nil(t, err)
				_, err = account.(wtypes.AccountPrivateKeyProvider).PrivateKey()
				assert.Nil(t, err)
			}
		})
	}
}

func TestAccountByNameDynamic(t *testing.T) {
	store := scratch.New()
	encryptor := keystorev4.New()
	wallet, err := hd.CreateWalletFromSeed("test wallet", []byte("wallet passphrase"), store, encryptor, []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	})
	require.NoError(t, err)

	err = wallet.Unlock([]byte("wallet passphrase"))
	require.NoError(t, err)

	account, err := wallet.AccountByName("m/12381/3600/0/0")
	require.NoError(t, err)
	assert.Equal(t, account.PublicKey().Marshal(), []byte{
		0xb9, 0x99, 0x79, 0x1a, 0x08, 0xa6, 0xa8, 0x45, 0xd8, 0x30, 0x4b, 0x79, 0x33, 0xc7, 0x3c, 0x3b,
		0x45, 0x89, 0x94, 0xcd, 0x27, 0x50, 0xbe, 0x4c, 0x77, 0x05, 0xed, 0x7a, 0xd2, 0xe2, 0x7e, 0xe1,
		0xc1, 0x44, 0xaf, 0xba, 0x9d, 0x1d, 0x81, 0x3d, 0x25, 0xec, 0x10, 0x35, 0x68, 0x84, 0x78, 0x0c,
	})
	account, err = wallet.AccountByName("m/12381/3600/1/1/1")
	require.NoError(t, err)
	assert.Equal(t, account.PublicKey().Marshal(), []byte{
		0x93, 0xdd, 0xa3, 0x6f, 0x78, 0x13, 0xa8, 0x8e, 0x30, 0xde, 0x60, 0x45, 0x0b, 0xfb, 0x0e, 0xd9,
		0x14, 0xc9, 0x04, 0x34, 0xdb, 0xa1, 0xbc, 0xdf, 0x1a, 0x63, 0x61, 0xa6, 0xf0, 0x7a, 0x78, 0xd2,
		0xfe, 0x38, 0x87, 0xc5, 0x82, 0xc3, 0xf4, 0x95, 0xcf, 0x6c, 0x8e, 0x01, 0xa4, 0x8d, 0xfd, 0xc5,
	})
}
