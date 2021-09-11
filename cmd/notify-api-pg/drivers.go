/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at https://mozilla.org/MPL/2.0/. */

package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/raohwork/envexist"
	"github.com/raohwork/notify"
	"github.com/raohwork/notify/drivers/httpdrv"
	"github.com/raohwork/notify/drivers/sendgriddrv"
	"github.com/raohwork/notify/drivers/smsav8d"
	"github.com/raohwork/notify/drivers/smtpdrv"
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
	keySMTPUser    = "SMTP_USER"
	keySMTPPass    = "SMTP_PASS"
	keySMTPServer  = "SMTP_SERVER"
	keySMTPAuth    = "SMTP_AUTH"
	keySMTPTLS     = "SMTP_TLS"
	keySMTPFrom    = "SMTP_FROM"
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
	m.May(keySMTPUser, "smtp user name", "user")
	m.May(keySMTPPass, "smtp password", "supersecret")
	m.May(keySMTPServer, "smtp server address (host:port)", "server:587")
	m.May(keySMTPTLS, "enable tls for smtp if not empty", "")
	m.May(keySMTPAuth, "smtp auth method, can be PLAIN/CRAMMD5 (case insensitive)", "plain")
	m.May(keySMTPFrom, "specify From header for smtp", "John Doe <john.doe@example.com>")
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

	smtpdrvs := initSMTP(data)
	dbdrv, err := pgsqldrv.New(db, drvs+len(smtpdrvs), int(thread))
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
	if len(smtpdrvs) > 0 {
		for _, d := range smtpdrvs {
			api.Register(d)
		}
		log.Printf("All SMTP settings are valid, SMTP drivers (plaintext and html) are enabled")
	}
}

func initSMTP(data map[string]string) (ret []types.Driver) {
	user, pass := data[keySMTPUser], data[keySMTPPass]
	if user == "" || pass == "" {
		return
	}
	server, auth := data[keySMTPServer], strings.ToLower(data[keySMTPAuth])
	if server == "" {
		return
	}
	arr := strings.Split(server, ":")
	var a smtp.Auth
	switch auth {
	case "plain":
		a = smtp.PlainAuth("", user, pass, arr[0])
	case "crammd5":
		a = smtp.CRAMMD5Auth(user, pass)
	default:
		return
	}
	useTLS := (data[keySMTPTLS] != "")
	f := data[keySMTPFrom]
	if f == "" {
		return
	}
	from, err := mail.ParseAddress(f)
	if err != nil {
		return
	}
	tlsHost := ""
	if useTLS {
		tlsHost = arr[0]
	}
	return []types.Driver{
		smtpdrv.New(*from, server, a, tlsHost),
		smtpdrv.NewHTML(*from, server, a, tlsHost),
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
