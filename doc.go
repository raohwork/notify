/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Package notify provides simple way to create notify server.
//
// Creating a server is easy as
//
//   1. Select and initialize a db driver (see ./model/mysqldrv)
//   2. Select and initialize few notify driver (see ./drivers/...)
//   3. Create a configuration (see SenderOptions)
//   4. Create a server with SenderOptions, and Register() your drivers
//   5. ListenAndServe(), enjoy it
package notify
