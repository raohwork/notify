Package notify provides simple way to create notify server.

[![PkgGoDev](https://pkg.go.dev/badge/github.com/raohwork/notify)](https://pkg.go.dev/github.com/raohwork/notify)

Creating a server is easy as

1. Select and initialize a db driver (see ./model/mysqldrv)
2. Select and initialize few notify drivers (see ./drivers/...)
3. Create a configuration (see SenderOptions)
4. Create a server with SenderOptions, and Register() your drivers
5. Start(), enjoy it

### Predefined binaries

There're two predefined binaries here. Both enables httpdrv.* drivers by default. You can enable sendgriddrv.* / tgdrv.* by setting environment variables.

The only difference between `notify-api` and `notify-api-pg` is database driver, fore one uses MySQL, the other uses PosgreSQL.

There're [automated built Docker images](https://hub.docker.com/repository/docker/raohwork/notify/tags?page=1) on docker hub. Tag `latest` is for `notify-api` and `pg` for `notify-api-pg`.


### FAQ

##### This is not well tested.

PR plz.

##### Is this safe to use in production?

For small scale, yes. I'm using it in production for few months, about few thousands of notifications per day, and few dozens per minute at peak.

Not tested in larger scale.

##### Can you add xxxx driver?

PR of driver that depends on external service (like mail driver with mailgun) will not be accepted, since I have no chance/spare time to test/maintain. Create your own repo and notify me via issue, I'll add a link in readme.

PR of DB driver is welcome if it can be tested using docker. See ./model/mysqldrv/mysql_test.go for example.

# License

MPL version 2.0
