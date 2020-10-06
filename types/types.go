/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package types defines common types used in package notify, model and drivers
package types

import (
	"encoding/json"
	"time"
)

//  State denotes the processing state of a notification
type State int

const (
	PENDING State = iota // notification is waiting to (re)send
	SUCCESS              // notification is sent to the endpoint
	FAILED               // notification is failed to send after all
)

// Scheduler is an user-defined function to determine when to resend notification
type Scheduler func(driver, notifyID string, lastExec time.Time, tried uint32) (next time.Time, stop bool)

// Params defines required parameters of API endpoint /send and /sendOnce
type Params struct {
	// notification ID. It has to be unique as used as primary key in DB.
	ID string `json:"id"`
	// driver type
	Driver string `json:"type"`
	// endpoint to recieve notification. see docs of the driver for detail
	Endpoint string `json:"endpoint"`
	// driver specific parameters. see docs of the driver for detail
	Payload json.RawMessage `json:"payload"`
}

// Driver defines the interface a driver must implement
type Driver interface {
	// driver type, 3rd party drivers *SHOULD* use go import path format like
	// github.com/some_org/some_proj/DRIVER_NAME
	Type() string
	// send the notification
	Send(ep string, content []byte) (resp []byte, err error)
	// check the payload format. content is in json format
	Verify(content []byte) (err error)
}

// Status defines response type of /status
type Status struct {
	CreateAt int64  `json:"create_at"`
	NextAt   int64  `json:"next_at"`
	Tried    uint32 `json:"tried"`
	State    State  `json:"state"`
}

// Detail defines response type of /detail
type Detail struct {
	Driver   string `json:"type"`
	Endpoint string `json:"endpoint"`
	Content  []byte `json:"content"`
	Response []byte `json:"response"`
	Status
}
