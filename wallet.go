// Copyright © 2019 - 2025 Weald Technology Trading.
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
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/wealdtech/go-ecodec"
	util "github.com/wealdtech/go-eth2-util"
	e2wtypes "github.com/wealdtech/go-eth2-wallet-types/v2"
	"github.com/wealdtech/go-indexer"
)

const (
	walletType = "hierarchical deterministic"
	version    = 1
)

// wallet contains the details of the wallet.
type wallet struct {
	id             uuid.UUID
	name           string
	version        uint
	crypto         map[string]any
	seed           []byte
	nextAccount    uint64
	store          e2wtypes.Store
	encryptor      e2wtypes.Encryptor
	index          *indexer.Index
	batch          *batch
	accounts       map[uuid.UUID]*account
	mutex          sync.Mutex
	creationMutex  sync.Mutex
	batchMutex     sync.Mutex
	batchDecrypted bool
}

// newWallet creates a new wallet.
func newWallet() *wallet {
	return &wallet{
		index:    indexer.New(),
		accounts: make(map[uuid.UUID]*account),
	}
}

// MarshalJSON implements custom JSON marshaller.
func (w *wallet) MarshalJSON() ([]byte, error) {
	data := make(map[string]any)
	data["uuid"] = w.id.String()
	data["name"] = w.name
	data["version"] = w.version
	data["type"] = walletType
	data["crypto"] = w.crypto
	data["nextaccount"] = w.nextAccount

	res, err := json.Marshal(data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal wallet")
	}

	return res, nil
}

// UnmarshalJSON implements custom JSON unmarshaller.
func (w *wallet) UnmarshalJSON(data []byte) error {
	var v map[string]any
	if err := json.Unmarshal(data, &v); err != nil {
		return errors.Wrap(err, "failed to unmarshal wallet")
	}

	if err := w.unmarshalType(v); err != nil {
		return err
	}
	if err := w.unmarshalID(v); err != nil {
		return err
	}
	if err := w.unmarshalName(v); err != nil {
		return err
	}
	if err := w.unmarshalCrypto(v); err != nil {
		return err
	}
	if err := w.unmarshalNextAccount(v); err != nil {
		return err
	}
	if err := w.unmarshalVersion(v); err != nil {
		return err
	}

	return nil
}

func (w *wallet) unmarshalType(v map[string]any) error {
	if val, exists := v["type"]; exists {
		dataWalletType, ok := val.(string)
		if !ok {
			return errors.New("wallet type invalid")
		}
		if dataWalletType != walletType {
			return fmt.Errorf("wallet type %q unexpected", dataWalletType)
		}
	} else {
		return errors.New("wallet type missing")
	}

	return nil
}

func (w *wallet) unmarshalID(v map[string]any) error {
	if val, exists := v["uuid"]; exists {
		idStr, ok := val.(string)
		if !ok {
			return errors.New("wallet ID invalid")
		}
		id, err := uuid.Parse(idStr)
		if err != nil {
			return errors.Wrap(err, "failed to parse UUID")
		}
		w.id = id
	} else {
		// Used to be ID; remove with V2.0
		if val, exists := v["id"]; exists {
			idStr, ok := val.(string)
			if !ok {
				return errors.New("wallet ID invalid")
			}
			id, err := uuid.Parse(idStr)
			if err != nil {
				return errors.Wrap(err, "failed to parse UUID")
			}
			w.id = id
		} else {
			return errors.New("wallet ID missing")
		}
	}

	return nil
}

func (w *wallet) unmarshalName(v map[string]any) error {
	if val, exists := v["name"]; exists {
		name, ok := val.(string)
		if !ok {
			return errors.New("wallet name invalid")
		}
		w.name = name
	} else {
		return errors.New("wallet name missing")
	}

	return nil
}

func (w *wallet) unmarshalCrypto(v map[string]any) error {
	if val, exists := v["crypto"]; exists {
		crypto, ok := val.(map[string]any)
		if !ok {
			return errors.New("wallet crypto invalid")
		}
		w.crypto = crypto
	} else {
		return errors.New("wallet crypto missing")
	}

	return nil
}

func (w *wallet) unmarshalNextAccount(v map[string]any) error {
	if val, exists := v["nextaccount"]; exists {
		nextAccount, ok := val.(float64)
		if !ok {
			return errors.New("wallet next account invalid")
		}
		w.nextAccount = uint64(nextAccount)
	} else {
		return errors.New("wallet next account missing")
	}

	return nil
}

func (w *wallet) unmarshalVersion(v map[string]any) error {
	if val, exists := v["version"]; exists {
		version, ok := val.(float64)
		if !ok {
			return errors.New("wallet version invalid")
		}
		w.version = uint(version)
	} else {
		return errors.New("wallet version missing")
	}

	return nil
}

// CreateWallet creates a wallet with the given name from a seed and stores it in the provided store.
func CreateWallet(ctx context.Context,
	name string,
	passphrase []byte,
	store e2wtypes.Store,
	encryptor e2wtypes.Encryptor,
	seed []byte,
) (
	e2wtypes.Wallet,
	error,
) {
	// First, try to open the wallet.
	_, err := OpenWallet(ctx, name, store, encryptor)
	if err == nil || !strings.Contains(err.Error(), "wallet not found") {
		return nil, fmt.Errorf("wallet %q already exists", name)
	}

	id, err := uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate UUID")
	}

	if len(seed) != 64 {
		return nil, errors.New("seed must be 64 bytes")
	}
	crypto, err := encryptor.Encrypt(seed, string(passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt seed")
	}

	// Decrypt to confirm it works.
	_, err = encryptor.Decrypt(crypto, string(passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt seed")
	}

	w := newWallet()
	w.id = id
	w.name = name
	w.crypto = crypto
	w.nextAccount = 0
	w.version = version
	w.store = store
	w.encryptor = encryptor

	if err := w.storeWallet(); err != nil {
		return nil, err
	}

	if err := w.storeAccountsIndex(); err != nil {
		return nil, err
	}

	return w, nil
}

// OpenWallet opens an existing wallet with the given name.
func OpenWallet(ctx context.Context, name string, store e2wtypes.Store, encryptor e2wtypes.Encryptor) (e2wtypes.Wallet, error) {
	data, err := store.RetrieveWallet(name)
	if err != nil {
		return nil, errors.Wrapf(err, "wallet %q does not exist", name)
	}

	return DeserializeWallet(ctx, data, store, encryptor)
}

// DeserializeWallet deserializes a wallet from its byte-level representation.
func DeserializeWallet(ctx context.Context,
	data []byte,
	store e2wtypes.Store,
	encryptor e2wtypes.Encryptor,
) (
	e2wtypes.Wallet,
	error,
) {
	wallet := newWallet()
	if err := json.Unmarshal(data, wallet); err != nil {
		return nil, errors.Wrap(err, "wallet corrupt")
	}
	wallet.store = store
	wallet.encryptor = encryptor
	if err := wallet.retrieveAccountsIndex(ctx); err != nil {
		return nil, errors.Wrap(err, "wallet index corrupt")
	}

	return wallet, nil
}

// ID provides the ID for the wallet.
func (w *wallet) ID() uuid.UUID {
	return w.id
}

// Type provides the type for the wallet.
func (w *wallet) Type() string {
	return walletType
}

// Name provides the name for the wallet.
func (w *wallet) Name() string {
	return w.name
}

// Version provides the version of the wallet.
func (w *wallet) Version() uint {
	return w.version
}

// store stores the wallet in the store.
func (w *wallet) storeWallet() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	data, err := json.Marshal(w)
	if err != nil {
		return errors.Wrap(err, "failed to marshal wallet")
	}

	if err := w.store.StoreWallet(w.ID(), w.Name(), data); err != nil {
		return errors.Wrap(err, "failed to store wallet")
	}

	return nil
}

// Lock locks the wallet.  A locked wallet cannot create new accounts.
func (w *wallet) Lock(_ context.Context) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.seed = nil

	return nil
}

// Unlock unlocks the wallet.  An unlocked wallet can create new accounts.
func (w *wallet) Unlock(_ context.Context, passphrase []byte) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	seed, err := w.encryptor.Decrypt(w.crypto, string(passphrase))
	if err != nil {
		return errors.New("incorrect passphrase")
	}
	w.seed = seed

	return nil
}

// IsUnlocked reports if the wallet is unlocked.
func (w *wallet) IsUnlocked(_ context.Context) (bool, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	return w.seed != nil, nil
}

// CreateAccount creates a new account in the wallet.
// The only rule for names is that they cannot start with an underscore (_) character.
func (w *wallet) CreateAccount(ctx context.Context, name string, passphrase []byte) (e2wtypes.Account, error) {
	// Lock the wallet to avoid parallel creation of accounts.
	w.creationMutex.Lock()
	defer w.creationMutex.Unlock()

	// Obtain the next available account.
	// Although this should be nextAccount, it is possible that the user has created wallets with explicit
	// paths that clash so we check here.
	accountNum := w.nextAccount
	var path string
	for {
		path = fmt.Sprintf("m/12381/3600/%d/0", accountNum)
		found := false
		for acc := range w.Accounts(ctx) {
			if acc.(*account).Path() == path {
				found = true
				break
			}
		}
		if !found {
			break
		}
		accountNum++
	}
	w.nextAccount = accountNum + 1

	if err := w.storeWallet(); err != nil {
		return nil, errors.Wrapf(err, "failed to update wallet for account %q", name)
	}

	account, err := w.createPathedAccount(ctx, path, name, passphrase)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// CreatePathedAccount creates a new account in the wallet with a given path.
// The only rule for names is that they cannot start with an underscore (_) character.
// This will error if an account with the name or path already exists.
func (w *wallet) CreatePathedAccount(ctx context.Context, path string, name string, passphrase []byte) (e2wtypes.Account, error) {
	// Lock the wallet to avoid parallel creation of accounts.
	w.creationMutex.Lock()
	defer w.creationMutex.Unlock()

	account, err := w.createPathedAccount(ctx, path, name, passphrase)
	if err != nil {
		return nil, err
	}

	return account, nil
}

// createPathedAccount creates a new account in the wallet with a given path.
// This is an internal function, that assumes a creation lock is held on the wallet.
func (w *wallet) createPathedAccount(ctx context.Context, path string, name string, passphrase []byte) (e2wtypes.Account, error) {
	if name == "" {
		return nil, errors.New("account name missing")
	}
	if strings.HasPrefix(name, "_") {
		return nil, fmt.Errorf("invalid account name %q", name)
	}
	if w.seed == nil {
		return nil, errors.New("wallet must be unlocked to create accounts")
	}

	// Ensure that we don't already have an account with this name.
	if _, err := w.AccountByName(ctx, name); err == nil {
		return nil, fmt.Errorf("account with name %q already exists", name)
	}

	// Ensure that we don't already have an account with this path.
	for acc := range w.Accounts(ctx) {
		if acc.(*account).Path() == path {
			return nil, fmt.Errorf("account with path %q already exists", path)
		}
	}
	// Generate the private key from the seed and next account
	privateKey, err := util.PrivateKeyFromSeedAndPath(w.seed, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create private key for account %q", name)
	}
	a := newAccount()
	a.path = path
	if a.id, err = uuid.NewRandom(); err != nil {
		return nil, errors.Wrap(err, "failed to generate UUID")
	}
	a.name = name
	a.publicKey = privateKey.PublicKey()
	// Encrypt the private key
	a.crypto, err = w.encryptor.Encrypt(privateKey.Marshal(), string(passphrase))
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt private key")
	}
	a.encryptor = w.encryptor
	a.version = w.encryptor.Version()
	a.wallet = w

	w.index.Add(a.id, a.name)

	if err := a.storeAccount(ctx); err != nil {
		return nil, err
	}

	if err := w.storeAccountsIndex(); err != nil {
		return nil, errors.Wrap(err, "failed to store account index")
	}

	return a, nil
}

func (w *wallet) retrieveBatchIfRequired(ctx context.Context) error {
	var err error

	if w.batch == nil {
		// Batch not retrieved, try to retrieve it now.
		if _, isBatchRetriever := w.store.(e2wtypes.BatchRetriever); isBatchRetriever {
			err = w.retrieveAccountsBatch(ctx)
		}
	}

	return err
}

// Accounts provides all accounts in the wallet.
func (w *wallet) Accounts(ctx context.Context) <-chan e2wtypes.Account {
	ch := make(chan e2wtypes.Account, 1024)

	go func(ch chan e2wtypes.Account) {
		_ = w.retrieveBatchIfRequired(ctx)

		if w.batch != nil && len(w.batch.entries) > 0 {
			// Batch present, use pre-loaded accounts.
			for _, account := range w.accounts {
				ch <- account
			}
			close(ch)

			return
		}

		// No batch; fall back to individual accounts on the store.
		for data := range w.store.RetrieveAccounts(w.ID()) {
			if account, err := deserializeAccount(w, data); err == nil {
				ch <- account
			}
		}
		close(ch)
	}(ch)

	return ch
}

// Export exports the entire wallet, protected by an additional passphrase.
func (w *wallet) Export(ctx context.Context, passphrase []byte) ([]byte, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	type walletExt struct {
		Wallet   *wallet    `json:"wallet"`
		Accounts []*account `json:"accounts"`
	}

	accounts := make([]*account, 0)
	for data := range w.store.RetrieveAccounts(w.ID()) {
		account, err := deserializeAccount(w, data)
		if err != nil {
			return nil, errors.Wrap(err, "failed to deserialize account")
		}
		accounts = append(accounts, account)
	}

	ext := &walletExt{
		Wallet:   w,
		Accounts: accounts,
	}

	data, err := json.Marshal(ext)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal wallet for export")
	}

	res, err := ecodec.Encrypt(data, passphrase)
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt export")
	}

	return res, nil
}

// Import imports the entire wallet, protected by an additional passphrase.
func Import(ctx context.Context,
	encryptedData []byte,
	passphrase []byte,
	store e2wtypes.Store,
	encryptor e2wtypes.Encryptor,
) (
	e2wtypes.Wallet,
	error,
) {
	type walletExt struct {
		Wallet   *wallet    `json:"wallet"`
		Accounts []*account `json:"accounts"`
	}

	data, err := ecodec.Decrypt(encryptedData, passphrase)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt wallet")
	}

	//  Create the wallet.
	ext := &walletExt{
		Wallet: newWallet(),
	}
	ext.Wallet.store = store
	ext.Wallet.encryptor = encryptor
	if err := json.Unmarshal(data, ext); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal wallet import")
	}

	// See if the wallet already exists.
	if _, err := OpenWallet(ctx, ext.Wallet.Name(), store, encryptor); err == nil {
		return nil, fmt.Errorf("wallet %q already exists", ext.Wallet.Name())
	}

	// Store the wallet.
	if err := ext.Wallet.storeWallet(); err != nil {
		return nil, errors.Wrapf(err, "failed to store wallet %q", ext.Wallet.Name())
	}

	// Create the accounts.
	for _, acc := range ext.Accounts {
		acc.wallet = ext.Wallet
		acc.encryptor = encryptor
		ext.Wallet.index.Add(acc.id, acc.name)
		if err := acc.storeAccount(ctx); err != nil {
			return nil, errors.Wrapf(err, "failed to store account %q", acc.Name())
		}
	}

	if err := ext.Wallet.storeAccountsIndex(); err != nil {
		return nil, errors.Wrap(err, "failed to store wallet index")
	}

	return ext.Wallet, nil
}

// AccountByName provides a single account from the wallet given its name.
// This will error if the account is not found.
func (w *wallet) AccountByName(ctx context.Context, name string) (e2wtypes.Account, error) {
	if strings.HasPrefix(name, "m/") {
		// Programmatic name
		return w.programmaticAccount(name)
	}

	id, exists := w.index.ID(name)
	if !exists {
		return nil, fmt.Errorf("no account with name %q", name)
	}

	return w.AccountByID(ctx, id)
}

// AccountByID provides a single account from the wallet given its ID.
// This will error if the account is not found.
func (w *wallet) AccountByID(ctx context.Context, id uuid.UUID) (e2wtypes.Account, error) {
	_ = w.retrieveBatchIfRequired(ctx)

	if w.batch != nil && len(w.batch.entries) > 0 {
		// Batch present, use pre-loaded account if available.
		if account, exists := w.accounts[id]; exists {
			return account, nil
		}
	}

	// No batch or account not in batch; fall back to individual account on the store.
	data, err := w.store.RetrieveAccount(w.id, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to retrieve account")
	}
	res, err := deserializeAccount(w, data)
	if err != nil {
		return nil, err
	}
	w.accounts[id] = res

	return res, nil
}

// Store returns the wallet's store.
func (w *wallet) Store() e2wtypes.Store {
	return w.store
}

// programmaticAccount calculates an account on the fly given its path.
func (w *wallet) programmaticAccount(path string) (e2wtypes.Account, error) {
	privateKey, err := util.PrivateKeyFromSeedAndPath(w.seed, path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create private key for path %q", path)
	}
	a := newAccount()
	a.path = path
	a.id, err = uuid.NewRandom()
	if err != nil {
		return nil, errors.Wrap(err, "failed to generate UUID")
	}
	a.name = path
	a.publicKey = privateKey.PublicKey()
	a.secretKey = privateKey
	// Encrypt the private key with an empty passphrase.
	a.crypto, err = w.encryptor.Encrypt(privateKey.Marshal(), "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to encrypt private key")
	}
	a.encryptor = w.encryptor
	a.version = w.encryptor.Version()
	a.wallet = w

	return a, nil
}

// retrieveAccountsIndex retrieves the accounts index for a wallet.
func (w *wallet) retrieveAccountsIndex(ctx context.Context) error {
	serializedIndex, err := w.store.RetrieveAccountsIndex(w.id)
	if err != nil {
		// Attempt to recreate the index.
		w.index = indexer.New()
		for account := range w.Accounts(ctx) {
			w.index.Add(account.ID(), account.Name())
		}
		if err := w.storeAccountsIndex(); err != nil {
			return err
		}
	} else {
		index, err := indexer.Deserialize(serializedIndex)
		if err != nil {
			return errors.Wrap(err, "failed to deserialize index")
		}
		w.index = index
	}

	return nil
}

// storeAccountsIndex stores the accounts index for a wallet.
func (w *wallet) storeAccountsIndex() error {
	serializedIndex, err := w.index.Serialize()
	if err != nil {
		return errors.Wrap(err, "failed to serialize index")
	}

	if err := w.store.StoreAccountsIndex(w.id, serializedIndex); err != nil {
		return errors.Wrap(err, "failed to store accounts index")
	}

	return nil
}
