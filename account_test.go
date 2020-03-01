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

	// Try to create without unlocking the wallet; should fail
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

				// Should not be able to obtain private key from a locked account
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
