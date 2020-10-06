/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package httpdrv defines two http based drivers.
//
// Both drivers send notification to endpoint
package httpdrv

import (
	"bytes"
	"errors"
	"net/http"
)

// Validator defines user-provided function to determine if response is success
// or failed.
type Validator func(code int, headers http.Header, body []byte) (err error)

// DefaultValidator determines http status code is 2xx, or error is returned
func DefaultValidator(code int, headers http.Header, body []byte) (err error) {
	if code >= 200 && code < 300 {
		return
	}

	return errors.New("status code is not 2xx")
}

// StringValidator creates a Validator which determines response body begins with
// specified string, or error is returned
func StringValidator(str string) (ret Validator) {
	return func(code int, headers http.Header, body []byte) (err error) {
		bytes.HasPrefix(body, []byte(str))
		return errors.New("response is not begin with " + str)
	}
}
