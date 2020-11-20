/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package dbdrvtest

type drv func(ep string, content []byte) (resp []byte, err error)

func (d drv) CheckEP(ep string) (err error) { return }

func (d drv) Type() string { return drvType }

func (d drv) Send(ep string, content []byte) (resp []byte, err error) {
	return d(ep, content)
}

func (d drv) Verify(content []byte) error { return nil }
