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

package hd

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	keystorev4 "github.com/wealdtech/go-eth2-wallet-encryptor-keystorev4"
)

type batchEntryJSON struct {
	UUID   uuid.UUID `json:"uuid"`
	Name   string    `json:"name"`
	Pubkey string    `json:"pubkey"`
	Path   string    `json:"path"`
}

func (b *batchEntry) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(&batchEntryJSON{
		UUID:   b.id,
		Name:   b.name,
		Pubkey: fmt.Sprintf("%#x", b.pubkey),
		Path:   b.path,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal JSON")
	}

	return res, nil
}

func (b *batchEntry) UnmarshalJSON(input []byte) error {
	data := batchEntryJSON{}
	if err := json.Unmarshal(input, &data); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}
	b.id = data.UUID
	b.name = data.Name
	var err error
	b.pubkey, err = hex.DecodeString(strings.TrimPrefix(data.Pubkey, "0x"))
	if err != nil {
		return errors.Wrap(err, "invalid pubkey")
	}
	b.path = data.Path

	return nil
}

type batchJSON struct {
	Entries   []*batchEntry  `json:"entries"`
	Crypto    map[string]any `json:"crypto"`
	Encryptor string         `json:"encryptor"`
	Version   int            `json:"version"`
}

func (b *batch) MarshalJSON() ([]byte, error) {
	res, err := json.Marshal(&batchJSON{
		Entries:   b.entries,
		Crypto:    b.crypto,
		Encryptor: b.encryptor.String(),
		Version:   version,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal JSON")
	}

	return res, nil
}

func (b *batch) UnmarshalJSON(input []byte) error {
	data := batchJSON{}
	if err := json.Unmarshal(input, &data); err != nil {
		return errors.Wrap(err, "invalid JSON")
	}
	if data.Version != version {
		return fmt.Errorf("unsupported version %d", data.Version)
	}
	b.entries = data.Entries
	switch data.Encryptor {
	case "keystorev4":
		b.encryptor = keystorev4.New()
	default:
		return fmt.Errorf("unsupported encryptor %s", data.Encryptor)
	}
	b.crypto = data.Crypto

	return nil
}
