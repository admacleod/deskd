[SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>]::
[SPDX-License-Identifier: AGPL-3.0-only]::

# Desk Booking

This is a fairly simple desk booking system. 
It is purposefully simple to make it easier to deploy.

The code is made available under the [AGPLv3](https://www.gnu.org/licenses/agpl-3.0.en.html).
For commercial licences please get in touch.

`deskd` is intended to be run as a cgi program on a web server.

## Building

Everything is Go, so you can just use `go build` or `go install` to build binaries.

## Configuration

`deskd` reads its configuration from the same places (in order of precedence):

- commandline flags passed directly to each tool
- environment variables, prefixed with `DESKD_`

The configuration options are as follows:

| Flag     | Env           | Type      | Default     | Description               |
|----------|---------------|-----------|-------------|---------------------------|
| `-db`    | `DESKD_DB`    | `string`  | `"test.db"` | Location of the database  |
| `-desks` | `DESKD_DESKS` | `string`  | `"desks"`   | Location of the desk file |

### Desk File

The desk file is used to define desks available for bookings.

It should be a single file with desk names defined, one per line.

Example:
```
FE1
FE2
# Comments don't work, this is now a bookable desk name.
FE3
SS1
```

**Warning:** removing a desk from the file does not remove associated bookings
from the database.

## Deployment

The intention is for `deskd` to be run on OpenBSD httpd, the details of how to
do so are documented below.

### Configure chroot

Because this is a Go program, and we cannot yet statically compile without CGo on
OpenBSD with sqlite you will need to add some libraries to the chroot:
```
mkdir -p /var/www/usr/lib
mkdir -p /var/www/usr/libexec
cp /usr/lib/libc.so.<version> /var/www/usr/lib/
cp /usr/lib/libpthread.so.<version> /var/www/usr/lib/
cp /usr/libexec/ld.so /var/www/usr/libexec/
```

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

If you would like deskd to use your favicon and display a floorplan when booking
desks then copy `favicon.ico` and `floorplan.png` into a suitable location, I
tend to choose `/var/www/htdocs/static`.

### Configure httpd

Update the httpd.conf to correctly use TLS (**NEVER** do HTTP Basic Auth over an
unsecured channel), to redirect to TLS and to serve CGI and static files.

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
			param DESKD_DB "/db/deskd.db"
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
