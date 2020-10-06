/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package model encapsules db operations. Only for internal purpose.
package model

import (
	"github.com/raohwork/notify/types"
)

type Item struct {
	ID       string
	Driver   string
	Endpoint string
	Content  []byte
	CreateAt int64
	NextAt   int64
	Tried    uint32
	State    types.State
}
