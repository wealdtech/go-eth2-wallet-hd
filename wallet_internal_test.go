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

package hd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	filesystem "github.com/wealdtech/go-eth2-wallet-store-filesystem"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

func TestUnmarshalWallet(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		err        error
		id         uuid.UUID
		version    uint
		walletType string
	}{
		{
			name: "Nil",
			err:  errors.New("unexpected end of JSON input"),
		},
		{
			name:  "Empty",
			input: []byte{},
			err:   errors.New("unexpected end of JSON input"),
		},
		{
			name:  "NotJSON",
			input: []byte(`bad`),
			err:   errors.New(`invalid character 'b' looking for beginning of value`),
		},
		{
			name:  "WrongJSON",
			input: []byte(`[]`),
			err:   errors.New(`json: cannot unmarshal array into Go value of type map[string]interface {}`),
		},
		{
			name:  "MissingID",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet ID missing"),
		},
		{
			name:  "WrongID",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":7,"name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet ID invalid"),
		},
		{
			name:  "BadID",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"wadbadba-dbad-badb-adba-badbadbadbad","name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("invalid UUID format"),
		},
		{
			name:  "WrongOldID",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"id":7,"name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet ID invalid"),
		},
		{
			name:  "BadOldID",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"id":"wadbadba-dbad-badb-adba-badbadbadbad","name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("invalid UUID format"),
		},
		{
			name:  "MissingName",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"id":"7603a428-999c-49d0-8241-ddfd63ee143d","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet name missing"),
		},
		{
			name:  "WrongName",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":2,"nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet name invalid"),
		},
		{
			name:  "MissingCrypto",
			input: []byte(`{"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet crypto missing"),
		},
		{
			name:  "WrongCrypto",
			input: []byte(`{"crypto":"foo","uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet crypto invalid"),
		},
		{
			name:  "MissingNextAccount",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet next account missing"),
		},
		{
			name:  "BadNextAccount",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":"bad","type":"hierarchical deterministic","version":1}`),
			err:   errors.New("wallet next account invalid"),
		},
		{
			name:  "MissingType",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"version":1}`),
			err:   errors.New("wallet type missing"),
		},
		{
			name:  "WrongType",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"type":2,"version":1}`),
			err:   errors.New("wallet type invalid"),
		},
		{
			name:  "BadType",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"type":"xd","version":1}`),
			err:   errors.New(`wallet type "xd" unexpected`),
		},
		{
			name:  "MissingVersion",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic"}`),
			err:   errors.New("wallet version missing"),
		},
		{
			name:  "WrongVersion",
			input: []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":false}`),
			err:   errors.New("wallet version invalid"),
		},
		{
			name:       "Good",
			input:      []byte(`{"crypto":{"checksum":{"function":"sha256","message":"d6f4c3898450a44666538785f419a78decde53da5f3ec17e611a961e204ed617","params":{}},"cipher":{"function":"aes-128-ctr","message":"0040872e1ba675bfe39053565f7ec02bc1560b2a95670b046f1a2e17facc1b57","params":{"iv":"7cbadf81a3895dbfee3863f0e5bd19f2"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"fcb4992215d5f84444c6f49a69e2124a899740e76caea09a1d465a71f802023a"}}},"uuid":"7603a428-999c-49d0-8241-ddfd63ee143d","name":"hd wallet","nextaccount":2,"type":"hierarchical deterministic","version":1}`),
			walletType: "hierarchical deterministic",
			id:         uuid.MustParse("7603a428-999c-49d0-8241-ddfd63ee143d"),
			version:    1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := newWallet()
			err := json.Unmarshal(test.input, output)
			if test.err != nil {
				require.NotNil(t, err)
				assert.Equal(t, test.err.Error(), err.Error())
			} else {
				require.Nil(t, err)
				assert.Equal(t, test.id, output.ID())
				assert.Equal(t, test.version, output.Version())
				assert.Equal(t, test.walletType, output.Type())
			}
		})
	}
}

func TestRetrieveAccountsIndex(t *testing.T) {
	rand.Seed(time.Now().Unix())
	// #nosec G404
	path := filepath.Join(os.TempDir(), fmt.Sprintf("TestRetrieveAccountsIndex-%d", rand.Int31()))
	defer os.RemoveAll(path)
	store := filesystem.New(filesystem.WithLocation(path))
	encryptor := keystorev4.New()
	seed := []byte{
		0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	}
	w, err := CreateWallet(context.Background(), "test wallet", []byte("pass"), store, encryptor, seed)
	require.NoError(t, err)
	require.NoError(t, w.(e2wtypes.WalletLocker).Unlock(context.Background(), []byte("pass")))

	account1, err := w.(e2wtypes.WalletAccountCreator).CreateAccount(context.Background(), "account1", []byte("test"))
	require.NoError(t, err)

	account2, err := w.(e2wtypes.WalletAccountCreator).CreateAccount(context.Background(), "account2", []byte("test"))
	require.NoError(t, err)

	idx, found := w.(*wallet).index.ID(account1.Name())
	require.True(t, found)
	require.Equal(t, account1.ID(), idx)

	idx, found = w.(*wallet).index.ID(account2.Name())
	require.True(t, found)
	require.Equal(t, account2.ID(), idx)

	_, found = w.(*wallet).index.ID("not present")
	require.False(t, found)

	// Manually delete the wallet index.
	indexPath := filepath.Join(path, w.ID().String(), "index")
	_, err = os.Stat(indexPath)
	require.False(t, os.IsNotExist(err))
	os.Remove(indexPath)
	_, err = os.Stat(indexPath)
	require.True(t, os.IsNotExist(err))

	// Re-open the wallet with a new store, to force re-creation of the index.
	store = filesystem.New(filesystem.WithLocation(path))
	w, err = OpenWallet(context.Background(), "test wallet", store, encryptor)
	require.NoError(t, err)

	require.NoError(t, w.(*wallet).retrieveAccountsIndex(context.Background()))
	idx, found = w.(*wallet).index.ID(account1.Name())
	require.True(t, found)
	require.Equal(t, account1.ID(), idx)

	idx, found = w.(*wallet).index.ID(account2.Name())
	require.True(t, found)
	require.Equal(t, account2.ID(), idx)

	_, found = w.(*wallet).index.ID("not present")
	require.False(t, found)
}
