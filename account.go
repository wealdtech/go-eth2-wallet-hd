// Copyright 2019 - 2023 Weald Technology Trading.
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
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	e2types "github.com/wealdtech/go-eth2-types/v2"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
)

// account contains the details of the account.
type account struct {
	id        uuid.UUID
	name      string
	publicKey e2types.PublicKey
	crypto    map[string]any
	unlocked  bool
	secretKey e2types.PrivateKey
	version   uint
	path      string
	wallet    *wallet
	encryptor e2wtypes.Encryptor
	mutex     sync.Mutex
}

// newAccount creates a new account
func newAccount() *account {
	return &account{}
}

// MarshalJSON implements custom JSON marshaller.
func (a *account) MarshalJSON() ([]byte, error) {
	data := make(map[string]any)
	data["uuid"] = a.id.String()
	data["name"] = a.name
	data["pubkey"] = fmt.Sprintf("%x", a.publicKey.Marshal())
	data["crypto"] = a.crypto
	data["path"] = a.path
	data["version"] = a.version
	return json.Marshal(data)
}

// UnmarshalJSON implements custom JSON unmarshaller.
func (a *account) UnmarshalJSON(data []byte) error {
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if val, exists := v["uuid"]; exists {
		idStr, ok := val.(string)
		if !ok {
			return errors.New("account ID invalid")
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return err
		}
		a.id = id
	} else {
		// Used to be ID; remove with V2.0
		if val, exists := v["id"]; exists {
			idStr, ok := val.(string)
			if !ok {
				return errors.New("account ID invalid")
			}
			id, err := uuid.Parse(idStr)
			if err != nil {
				return err
			}
			a.id = id
		} else {
			return errors.New("account ID missing")
		}
	}
	if val, exists := v["name"]; exists {
		name, ok := val.(string)
		if !ok {
			return errors.New("account name invalid")
		}
		a.name = name
	} else {
		return errors.New("account name missing")
	}
	if val, exists := v["pubkey"]; exists {
		publicKey, ok := val.(string)
		if !ok {
			return errors.New("account pubkey invalid")
		}
		bytes, err := hex.DecodeString(publicKey)
		if err != nil {
			return err
		}
		a.publicKey, err = e2types.BLSPublicKeyFromBytes(bytes)
		if err != nil {
			return err
		}
	} else {
		return errors.New("account pubkey missing")
	}
	if val, exists := v["crypto"]; exists {
		crypto, ok := val.(map[string]any)
		if !ok {
			return errors.New("account crypto invalid")
		}
		a.crypto = crypto
	} else {
		return errors.New("account crypto missing")
	}
	if val, exists := v["path"]; exists {
		path, ok := val.(string)
		if !ok {
			return errors.New("account path invalid")
		}
		a.path = path
	} else {
		return errors.New("account path missing")
	}
	if val, exists := v["version"]; exists {
		version, ok := val.(float64)
		if !ok {
			return errors.New("account version invalid")
		}
		a.version = uint(version)
	} else {
		return errors.New("account version missing")
	}
	// Only support keystorev4 at current...
	if a.version == 4 {
		a.encryptor = keystorev4.New()
	} else {
		return errors.New("unsupported keystore version")
	}

	return nil
}

// ID provides the ID for the account.
func (a *account) ID() uuid.UUID {
	return a.id
}

// Name provides the ID for the account.
func (a *account) Name() string {
	return a.name
}

// PublicKey provides the public key for the account.
func (a *account) PublicKey() e2types.PublicKey {
	// Safe to ignore the error as this is already a public key
	keyCopy, _ := e2types.BLSPublicKeyFromBytes(a.publicKey.Marshal())
	return keyCopy
}

// PrivateKey provides the private key for the account.
func (a *account) PrivateKey(ctx context.Context) (e2types.PrivateKey, error) {
	if !a.unlocked {
		return nil, errors.New("cannot provide private key when account is locked")
	}

	return e2types.BLSPrivateKeyFromBytes(a.secretKey.Marshal())
}

// Wallet provides the wallet for the account.
func (a *account) Wallet() e2wtypes.Wallet {
	return a.wallet
}

// Lock locks the account.  A locked account cannot sign data.
func (a *account) Lock(ctx context.Context) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.unlocked = false

	return nil
}

// Unlock unlocks the account.  An unlocked account can sign data.
func (a *account) Unlock(ctx context.Context, passphrase []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if a.unlocked {
		// The account is already unlocked; nothing to do.
		return nil
	}

	if a.secretKey == nil {
		// First time unlocking, need to decrypt the secret key.
		if a.crypto == nil {
			// This is a batch account, decrypt the batch.
			if err := a.wallet.batchDecrypt(ctx, passphrase); err != nil {
				return errors.New("incorrect batch passphrase")
			}
		} else {
			// This is an individual account, decrypt the account.
			secretKeyBytes, err := a.encryptor.Decrypt(a.crypto, string(passphrase))
			if err != nil {
				return errors.New("incorrect passphrase")
			}
			secretKey, err := e2types.BLSPrivateKeyFromBytes(secretKeyBytes)
			if err != nil {
				return errors.Wrap(err, "failed to obtain private key")
			}
			a.secretKey = secretKey
		}

		// Ensure the private key is correct.
		publicKey := a.secretKey.PublicKey()
		if !bytes.Equal(publicKey.Marshal(), a.publicKey.Marshal()) {
			a.secretKey = nil
			return errors.New("private key does not correspond to public key")
		}
	}
	a.unlocked = true

	return nil
}

// IsUnlocked returns true if the account is unlocked.
func (a *account) IsUnlocked(ctx context.Context) (bool, error) {
	return a.unlocked, nil
}

// Path returns the full path from which the account key is derived.
func (a *account) Path() string {
	return a.path
}

// Sign signs data.
func (a *account) Sign(_ context.Context, data []byte) (e2types.Signature, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if !a.unlocked {
		return nil, errors.New("cannot sign when account is locked")
	}

	return a.secretKey.Sign(data), nil
}

// storeAccount stores the account.
func (a *account) storeAccount(ctx context.Context) error {
	data, err := json.Marshal(a)
	if err != nil {
		return errors.Wrap(err, "failed to create store format")
	}

	if err := a.wallet.store.StoreAccount(a.wallet.ID(), a.ID(), data); err != nil {
		return errors.Wrap(err, "failed to store account")
	}

	// Check to ensure the account can be retrieved.
	if _, err = a.wallet.AccountByName(ctx, a.name); err != nil {
		return errors.Wrap(err, "failed to confirm account when retrieving by name")
	}
	if _, err = a.wallet.AccountByID(ctx, a.id); err != nil {
		return errors.Wrap(err, "failed to confirm account when retrieveing by ID")
	}

	return nil
}

// deserializeAccount deserializes account data to an account.
func deserializeAccount(w *wallet, data []byte) (*account, error) {
	a := newAccount()
	a.wallet = w
	a.encryptor = w.encryptor
	if err := json.Unmarshal(data, a); err != nil {
		return nil, err
	}
	return a, nil
}
