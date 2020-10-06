/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

// Command notify-api is predefined binary with provided drivers using MySQL.
//
// All drivers are configured via environment variables. Run "notify-api -h" for
// detail.
//
// Provided drivers
//
// httpdrv.HTTPGet/httpdrv.HTTPPost is always enabled. It uses
// httpdrv.StringValidator to parse response, with user-defined timeout.
//
// tgdrv.Markdown is enabled only when you properly configured token and target.
//
// sendgriddrv.New is enabled if you set api key.
package main

import (
	"context"
	"flag"
	"os"
	"os/signal"

	"github.com/raohwork/envexist"
)

func main() {
	if !envexist.Parse() {
		envexist.PrintEnvList()
		return
	}
	var help bool
	flag.BoolVar(&help, "h", false, "print envvars")
	flag.Parse()
	if help {
		envexist.PrintEnvList()
		return
	}

	api.GetHTTPServer().Addr = bind

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt)

		for range c {
			api.Shutdown(context.Background())
			return
		}
	}()

	api.Start()
}
