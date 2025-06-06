# Desk Booking

This is a fairly simple desk booking system. 
It is purposefully simple to make it easier to deploy.

The code is made available under the [AGPLv3](https://www.gnu.org/licenses/agpl-3.0.en.html).
For commercial licences please get in touch.

`deskd` is intended to be run as a cgi program on a web server.

## Building

Everything is Go, so you can just use `go build` or `go install` to build binaries.

It is now possible (and strongly recommended) to statically build this like so:
```
CGO_ENABLED=1 go build -ldflags '-s -w -linkmode external -extldflags "-fno-PIC -static"' -o deskd
```
so that no libraries need to be copied around if running in chroot or scratch containers.

## Configuration

`deskd` reads its configuration from the same places (in order of precedence):

- commandline flags passed directly to each tool
- environment variables, prefixed with `DESKD_`

The configuration options are as follows:

| Flag     | Env           | Type      | Default      | Description                        |
|----------|---------------|-----------|--------------|------------------------------------|
| `-db`    | `DESKD_DB`    | `string`  | `"db/deskd"` | Location of the database directory |

### Adding desks

As `deskd` uses a the filesystem to store bookings you can add and remove bookable desks by
adding and removing directories with the names of the desks from the top level of the database
directory.

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
