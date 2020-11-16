/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/raohwork/envexist"
	"github.com/raohwork/notify"
	"github.com/raohwork/notify/drivers/httpdrv"
	"github.com/raohwork/notify/drivers/sendgriddrv"
	"github.com/raohwork/notify/drivers/smsav8d"
	"github.com/raohwork/notify/drivers/tgdrv"
	"github.com/raohwork/notify/model"
	"github.com/raohwork/notify/model/pgsqldrv"
	"github.com/raohwork/notify/types"
)

var api notify.APIServer

const (
	keyHTTPBind    = "HTTP_BIND"
	keyHTTPTimeout = "HTTP_TIMEOUT"
	keyHTTPString  = "HTTP_STRING"
	keyTGToken     = "TG_TOKEN"
	keyTGTarget    = "TG_TARGET"
	keySendgridKey = "SENDGRID_KEY"
	keyDSN         = "DSN"
	keyMaxTry      = "MAX_TRY"
	keyThreads     = "THREADS"
	keyAV8DUser    = "AV8D_USER"
	keyAV8DPass    = "AV8D_PASS"
)

var bind string
var dbdrv model.DBDrv

func init() {
	m := envexist.New("NOTIFY", setup)
	m.Need(keyDSN, "pgsql connection string", "")
	m.May(keyHTTPTimeout, "httpdrv request timeout", "10")
	m.Want(keyTGToken, "telegram token", "")
	m.Want(keyTGTarget, "telegram targets", "chan1=-100123781523,chan2=-100129386128736")
	m.Want(keySendgridKey, "sendgrid api key", "")
	m.Want(keyAV8DUser, "user name of every8d", "")
	m.Want(keyAV8DPass, "password of every8d", "")
	m.May(keyHTTPBind, "api server bind address", ":8080")
	m.May(keyMaxTry, "retry at most these times", "6")
	m.May(keyThreads, "goroutines to send notification", "10")
	m.May(keyHTTPString, "string pass to httpdrv.StringValidator", "0000")
}

func setup(data map[string]string) {
	bind = data[keyHTTPBind]
	dsn := data[keyDSN]
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatal("cannot connect to db: ", err)
	}

	max, err := strconv.ParseUint(data[keyMaxTry], 10, 32)
	if err != nil || max == 0 {
		log.Fatal("MAX_TRY must be positive integer")
	}

	thread, err := strconv.ParseUint(data[keyThreads], 10, 16)
	if err != nil || thread == 0 {
		log.Fatal("THREAD must be positive integer")
	}

	t := time.Duration(15)
	if str := data[keyHTTPTimeout]; str != "" {
		x, e := strconv.ParseUint(str, 10, 64)
		if e == nil {
			t = time.Duration(x)
		}
	}
	t *= time.Second
	cl := &http.Client{
		Timeout: time.Duration(t) * time.Second,
	}

	drvs := 2
	x := make([]types.Driver, 0, 2)
	if token, target := data[keyTGToken], data[keyTGTarget]; token != "" && target != "" {
		d := initTG(token, target, cl)
		if l := len(d); l > 0 {
			drvs += l
			x = append(x, d...)
		}
	}

	if u, p := data[keyAV8DUser], data[keyAV8DPass]; u != "" && p != "" {
		log.Print("got username and password, enables smsav8d")
		d := smsav8d.New(u, p, cl)
		drvs++
		x = append(x, d)
	}

	if sg := data[keySendgridKey]; sg != "" {
		d := initSendgrid(sg, cl)
		if d != nil {
			drvs++
			x = append(x, d)
		}
	}

	dbdrv, err := pgsqldrv.New(db, drvs, int(thread))
	if err != nil {
		log.Fatal("cannot initialize db driver: ", err)
	}

	api, err = notify.NewAPI(notify.SenderOptions{
		MaxTries:   uint32(max),
		MaxThreads: uint16(thread),
		DBDrv:      dbdrv,
	})
	if err != nil {
		log.Fatal("cannot initialize api server: ", err)
	}

	api.Register(httpdrv.HTTPGet(
		cl,
		httpdrv.StringValidator(data[keyHTTPString]),
	))
	api.Register(httpdrv.HTTPPost(
		cl,
		httpdrv.StringValidator(data[keyHTTPString]),
	))
	log.Printf("HTTP callback is considered as success if response begins with %s", data[keyHTTPString])
	for _, d := range x {
		api.Register(d)
	}
}

func initSendgrid(key string, cl *http.Client) (ret types.Driver) {
	key = strings.TrimSpace(key)
	if len(key) == 0 {
		return
	}

	log.Print("sendgrid key detected, enabling sendgriddrv")
	return sendgriddrv.New(key, cl)
}

func initTG(token, target string, cl *http.Client) (ret []types.Driver) {
	token = strings.TrimSpace(token)
	if len(token) == 0 {
		return
	}

	targets := strings.Split(target, ",")
	m := map[string]int64{}
	for _, line := range targets {
		arr := strings.Split(line, "=")
		if len(arr) != 2 {
			log.Fatal("invalid tg targets")
		}
		k, v := strings.TrimSpace(arr[0]), strings.TrimSpace(arr[1])
		if k == "" || v == "" {
			log.Fatal("invalid tg targets")
		}

		i, e := strconv.ParseInt(v, 10, 64)
		if e != nil {
			log.Fatal("invalid tg targets")
		}

		m[k] = i
	}

	log.Print("telegram token and target detected, enabling tgdrv")
	x := make([]types.Driver, 3)

	d, err := tgdrv.Markdown(token, m, cl)
	if err != nil {
		log.Fatal("cannot init telegram driver: ", err)
	}
	x[0] = d

	d, err = tgdrv.HTML(token, m, cl)
	if err != nil {
		log.Fatal("cannot init telegram driver: ", err)
	}
	x[1] = d

	d, err = tgdrv.Plain(token, m, cl)
	if err != nil {
		log.Fatal("cannot init telegram driver: ", err)
	}
	x[2] = d

	return x
}
