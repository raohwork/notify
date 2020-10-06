/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package model

import (
	"time"

	"github.com/raohwork/notify/types"
)

type E404 struct{}

func (e *E404) Error() string { return "record not found" }

// DBDrv defines db related methods
//
// It is possible to do some magic in this interface to affect sender, but you
// *SHOULD NOT* do this unless you have good reason.
type DBDrv interface {
	// creates a notification to send
	Create(i *Item) (err error)
	// send a notification again, does not retry, return &E404{} if id not found
	Resend(id string, max uint32) (err error)
	// update a notification after sending. *NEVER* return error if id not found
	Update(id string, tried uint32, next int64, state types.State, resp []byte) (err error)
	// retrieve last sending result, return &E404{} if id not found
	Result(id string) (ret []byte, err error)
	// retrieve status, return &E404{} if id not found
	Status(id string) (ret types.Status, err error)
	// retrieve detail info, return &E404{} if id not found
	Detail(id string) (ret types.Detail, err error)
	// get one pending notification
	Pending(now int64, max uint32, drvs, ids []string) (ret *Item, err error)
	// delete a notification, excepts current sending notifications
	// *NEVER* return error if nothing's deleted (id not found or something)
	Delete(id string, cur []string) (err error)
	// clear finished notifications older than t, excepts current sending ones
	Clear(t time.Time, cur []string) (err error)
	// clear all notifications older than t, excepts current sending ones
	ForceClear(t time.Time, cur []string) (err error)
}
