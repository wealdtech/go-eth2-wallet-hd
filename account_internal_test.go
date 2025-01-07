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
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	e2types "github.com/wealdtech/go-eth2-types/v2"
)

func TestMain(m *testing.M) {
	if err := e2types.InitBLS(); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestUnmarshalAccount(t *testing.T) {
	tests := []struct {
		name       string
		input      []byte
		err        error
		id         uuid.UUID
		version    uint
		walletType string
		publicKey  []byte
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
			name:  "Blank",
			input: []byte(""),
			err:   errors.New("unexpected end of JSON input"),
		},
		{
			name:  "NotJSON",
			input: []byte(`bad`),
			err:   errors.New(`invalid character 'b' looking for beginning of value`),
		},
		{
			name:  "MissingID",
			input: []byte(`{"name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account ID missing"),
		},
		{
			name:  "WrongID",
			input: []byte(`{"uuid":1,"name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account ID invalid"),
		},
		{
			name:  "BadID",
			input: []byte(`{"uuid":"c99","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("failed to parse UUID: invalid UUID length: 3"),
		},
		{
			name:  "WrongOldID",
			input: []byte(`{"id":1,"name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account ID invalid"),
		},
		{
			name:  "BadOldID",
			input: []byte(`{"id":"c99","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("failed to parse UUID: invalid UUID length: 3"),
		},
		{
			name:  "WrongName",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":true,"pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account name invalid"),
		},
		{
			name:  "MissingCrypto",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account crypto missing"),
		},
		{
			name:  "BadCrypto",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":2,"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account crypto invalid"),
		},
		{
			name:  "MissingPath",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}}}`),
			err:   errors.New("account path missing"),
		},
		{
			name:  "BadPath",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":4}`),
			err:   errors.New("account path invalid"),
		},
		{
			name:  "MissingPubKey",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account pubkey missing"),
		},
		{
			name:  "InvalidPubKey",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":true,"version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New("account pubkey invalid"),
		},
		{
			name:  "BadPubKey",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44h","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New(`failed to decode public key: encoding/hex: invalid byte: U+0068 'h'`),
		},
		{
			name:  "BadPubKey2",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c4c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New(`account pubkey could not be decoded: public key must be 48 bytes`),
		},
		{
			name:  "MissingVersion",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New(`account version missing`),
		},
		{
			name:  "BadVersion",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":true,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New(`account version invalid`),
		},
		{
			name:  "WrongVersion",
			input: []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":3,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			err:   errors.New(`unsupported keystore version`),
		},
		{
			name:       "Keystore",
			input:      []byte(`{"crypto": {"kdf": {"function": "scrypt", "params": {"dklen": 32, "n": 262144, "r": 8, "p": 1, "salt": "1653ea66b5867a30919506bdfd247767ada85264700fc499d988761d26a33a15"}, "message": ""}, "checksum": {"function": "sha256", "params": {}, "message": "e4c993187b97b105b1ab44688d46350e6bb4b91369b878f851f87ba1187b829d"}, "cipher": {"function": "aes-128-ctr", "params": {"iv": "4c69b5c90d7b26a6972a481d19f8bbec"}, "message": "c8270de22051c2872c4162f0410515854b446ed11aba69d1f87acd4e3e1fe7cc"}}, "description": "", "pubkey": "a431d185cdc09056e33f9ac404021d58a01ffc55d2c4daf5b6ad85848d8d3ca501f775d9da585025a0ecbd902b87b2af", "path": "m/12381/3600/0/0/0", "uuid": "7759b20b-6956-4faa-b60d-33ebffad8f4d", "version": 4}`),
			walletType: "hierarchical deterministic",
			id:         uuid.MustParse("7759b20b-6956-4faa-b60d-33ebffad8f4d"),
			publicKey:  []byte{0xa4, 0x31, 0xd1, 0x85, 0xcd, 0xc0, 0x90, 0x56, 0xe3, 0x3f, 0x9a, 0xc4, 0x04, 0x02, 0x1d, 0x58, 0xa0, 0x1f, 0xfc, 0x55, 0xd2, 0xc4, 0xda, 0xf5, 0xb6, 0xad, 0x85, 0x84, 0x8d, 0x8d, 0x3c, 0xa5, 0x01, 0xf7, 0x75, 0xd9, 0xda, 0x58, 0x50, 0x25, 0xa0, 0xec, 0xbd, 0x90, 0x2b, 0x87, 0xb2, 0xaf},
			version:    4,
		},
		{
			name:       "Good",
			input:      []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			walletType: "hierarchical deterministic",
			id:         uuid.MustParse("c9958061-63d4-4a80-bcf3-25f3dda22340"),
			publicKey:  []byte{0xa9, 0x9a, 0x76, 0xed, 0x77, 0x96, 0xf7, 0xbe, 0x22, 0xd5, 0xb7, 0xe8, 0x5d, 0xee, 0xb7, 0xc5, 0x67, 0x7e, 0x88, 0xe5, 0x11, 0xe0, 0xb3, 0x37, 0x61, 0x8f, 0x8c, 0x4e, 0xb6, 0x13, 0x49, 0xb4, 0xbf, 0x2d, 0x15, 0x3f, 0x64, 0x9f, 0x7b, 0x53, 0x35, 0x9f, 0xe8, 0xb9, 0x4a, 0x38, 0xe4, 0x4c},
			version:    4,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			output := newAccount()
			err := json.Unmarshal(test.input, output)
			if test.err != nil {
				require.NotNil(t, err)
				assert.Equal(t, test.err.Error(), err.Error())
			} else {
				require.Nil(t, err)
				assert.Equal(t, test.id, output.ID())
				assert.Equal(t, test.publicKey, output.PublicKey().Marshal())
				//				assert.Equal(t, test.version, output.Version())
				//				assert.Equal(t, test.walletType, output.Type())
			}
		})
	}
}

func TestUnlock(t *testing.T) {
	tests := []struct {
		name       string
		account    []byte
		passphrase []byte
		err        error
	}{
		{
			name:       "PublicKeyMismatch",
			account:    []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"b89bebc699769726a318c8e9971bd3171297c61aea4a6578a7a4f94b547dcba5bac16a89108b6b6a1fe3695d1a874a0b","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			passphrase: []byte("test passphrase"),
			err:        errors.New("private key does not correspond to public key"),
		},
		{
			name:       "Keystore",
			account:    []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			passphrase: []byte("test passphrase"),
		},
		{
			name:       "BadPassphrase",
			account:    []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"09b65fda487a021900003a8b2081694b15ca73e0e59a5c79a5126f6818a2f171","params":{}},"cipher":{"function":"aes-128-ctr","message":"8386db98fbe002c02de9bc122b7680078045bf6c5c9ac2f7e8b53afbea0d3e15","params":{"iv":"45092570c625ad5e8decfcd991464740"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"ae6433afd822e6d99dfaa1a0d73d2ee263efdf62f858ba0c422cf27982d09c8a"}}},"path":"m/12381/3600/0/0"}`),
			passphrase: []byte("wrong passphrase"),
			err:        errors.New("incorrect passphrase"),
		},
		{
			name:       "EmptyPassphrase",
			account:    []byte(`{"uuid":"c9958061-63d4-4a80-bcf3-25f3dda22340","name":"test account","pubkey":"a99a76ed7796f7be22d5b7e85deeb7c5677e88e511e0b337618f8c4eb61349b4bf2d153f649f7b53359fe8b94a38e44c","version":4,"crypto":{"checksum":{"function":"sha256","message":"4a67cc6a4ff5e81235393c677652213cc96488d68f17d045f99f9cef8acc81a1","params":{}},"cipher":{"function":"aes-128-ctr","message":"ce7c1d11cd71adb604c055a2d198336387e0579275c4d2d45c184ed54631ebdd","params":{"iv":"c752efc43ca0651bb06adccf4b8651b8"}},"kdf":{"function":"pbkdf2","message":"","params":{"c":16,"dklen":32,"prf":"hmac-sha256","salt":"b49107e74e59a80ce5ac1624e6d27e7305aa22f5ffba4f602dd4dfe34fdf8640"}}},"path":"m/12381/3600/0/0"}`),
			passphrase: []byte(""),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			account := newAccount()
			err := json.Unmarshal(test.account, account)
			require.Nil(t, err)

			// Try to sign something - should fail because locked
			_, err = account.Sign(context.Background(), []byte("test"))
			assert.NotNil(t, err)

			err = account.Unlock(context.Background(), test.passphrase)
			if test.err != nil {
				require.NotNil(t, err)
				assert.Equal(t, test.err.Error(), err.Error())
			} else {
				require.Nil(t, err)

				// Try to sign something - should succeed because unlocked
				signature, err := account.Sign(context.Background(), []byte("test"))
				assert.Nil(t, err)

				verified := signature.Verify([]byte("test"), account.PublicKey())
				require.Nil(t, err)
				assert.Equal(t, true, verified)

				require.NoError(t, account.Lock(context.Background()))

				// Try to sign something - should fail because locked (again)
				_, err = account.Sign(context.Background(), []byte("test"))
				assert.NotNil(t, err)
			}
		})
	}
}
