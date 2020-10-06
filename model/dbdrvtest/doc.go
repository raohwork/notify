/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package dbdrvtest provides needed integral test for db driver writer.
//
// DBDrv implementatons should provide docker-based one-time storage configuration
// in it's "testdata" folder, and use NewSuite() and Run() to run these tests.
//
// The Suite will create an notify.APIServer at given address with special designed
// notify driver.
//
// See ../mysqldrv/mysql_test.go for example.
package dbdrvtest
