# Desk Booking

This is a fairly simple desk booking system. 
It is purposefully simple to make it easier to deploy.

The code is made available under the [AGPLv3](https://www.gnu.org/licenses/agpl-3.0.en.html).
For commercial licences please get in touch.

`deskd` is intended to be run as a cgi program on a web server.

Developer documentation is mostly in [`doc.go`](./doc.go) so you can read it using GoDoc.

## Building

Everything is Go, so you can just use `go build` or `go install` to build binaries.

It is now possible (and strongly recommended) to statically build this like so:
```
CGO_ENABLED=1 go build -tags 'sqlite_foreign_keys' -ldflags '-s -w -linkmode external -extldflags "-fno-PIC -static"' -o deskd
```
so that no libraries need to be copied around if running in chroot or scratch containers.

## Configuration

`deskd` reads its configuration from the environment only.

The configuration options are as follows:

| Env        | Type     | Default                                               | Description                                                |
|------------|----------|-------------------------------------------------------|------------------------------------------------------------|
| `DESKD_DB` | `string` | `"file:/db/deskd.db?cache=shared&_foreign_keys=true"` | The DSN used to access a sqlite database storing bookings. |

### Adding desks

As `deskd` uses the database to store bookings you can add and remove bookable desks by
adding and removing entries in the `desks` table, changes will be reflected in the
application immediately.

## Deployment

The intention is for `deskd` to be run on OpenBSD httpd, the details of how to
do so are documented below.

### Configure chroot

I often tend to create a separate directory for the db to live in, but make sure
it is writable by the `www` user:
```
mkdir -p /var/www/db
chown www:www /var/www/db
```

### Move the binary to the chroot

Build the binary and then move it to `/var/www/cgi-bin`.

### Configure htpasswd

You want user access security right?

Use `htpasswd(1)` to configure users to a `.htpasswd` file within the chroot
(`/var/www`).

**Note** the username you use for `.htpasswd` is also the username that will be
displayed inside the application to identify booked desks.

### Copy static files

I would advise you use the provided `style.css` to ensure users get something
a little nicer than totally bare HTML (although, feel free to write your own),
so copy this file into a suitable location, I tend to choose
`/var/www/htdocs/static`.

If you would like deskd to use your favicon and display a floorplan when booking
desks then copy `favicon.ico` and `floorplan.png` into the same location, and it 
will use them from there.

### Configure httpd

Update the httpd.conf to correctly use TLS (**NEVER** do HTTP Basic Auth over an
unsecured channel as it will leak usernames and passwords to anyone on the network),
to redirect to TLS, and to serve CGI and static files.

```
server "deskd.example.com" {
	listen on * tls port https
	authenticate with "/.htpasswd"
	tls {
		certificate "/etc/ssl/deskd.example.com.pem"
		key "/etc/ssl/private/deskd.example.com.key"
	}
	location "/static/*" {
		root "/htdocs"
	}
	location "*" {
		fastcgi {
			param SCRIPT_FILENAME "/cgi-bin/deskd"
			param DESKD_DB "/db/deskd"
		}
	}
}

server "deskd.example.com" {
	listen on * port http
	location "/.well-known/acme-challenge/*" {
		root "/acme"
		request strip 2
	}
	location "*" {
		block return 301 "https://$HTTP_HOST$REQUEST_URI"
	}
}
```

### Start everything

Start (or restart) the relevant services:
```
rcctl start httpd slowcgi
```

### Troubleshooting

You may have some permissions issues with the initial database setup (I need to
fix this), just make sure it can be read and written by `www` user.
