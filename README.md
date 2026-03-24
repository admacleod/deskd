# Desk Booking

This is a fairly simple desk booking system.
It is purposefully simple to make it easier to deploy.

The code is made available under the [AGPLv3](https://www.gnu.org/licenses/agpl-3.0.en.html).
For commercial licences please get in touch.

`deskd` is intended to be run as a CGI program on a web server.

## Building

`deskd` is written in C and only requires `sqlite3` as an external dependency.
On OpenBSD, install the sqlite3 package and then build with make:
```
pkg_add sqlite3
make
```

The build uses `cc` with `-std=c23` and links against `-lsqlite3`.
All other dependencies are satisfied by the C libraries available in OpenBSD base.

The binary is statically linked by default so it can run inside an OpenBSD
httpd chroot without access to shared libraries.

On macOS (for development and testing), static linking is not supported.
Install sqlite3 via MacPorts and build with static linking disabled:
```
port install sqlite3
make LDFLAGS_STATIC= CFLAGS="-I/opt/local/include" LDFLAGS="-L/opt/local/lib -lsqlite3"
```

A `compat.h` header provides shims for OpenBSD-specific functions (`reallocarray`,
`pledge`, `unveil`) so that the code compiles on other platforms without reducing
the safety of the OpenBSD build.

### Database migration

Before first use, run the migration to create the database schema:
```
./deskd migrate
```

This is safe to run repeatedly as it uses `CREATE TABLE IF NOT EXISTS`.

## Configuration

`deskd` reads its configuration from the environment only.

The configuration options are as follows:

| Env        | Type     | Description                                                |
|------------|----------|------------------------------------------------------------|
| `DESKD_DB` | `string` | **Required.** The DSN used to access a sqlite database storing bookings. |

The parent directory of the database file must already exist and be writable
by the application. `deskd` will not create directories automatically.

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

# Development Documentation

`deskd` is laid out so that each route handler is as independent as possible.
Rather than creating shared objects to handle all the required functionality,
each path can be considered as a totally separate script or application.

The rationale behind this separation is that this software is intended to run
as a CGI script; this means that every request results in a new invocation of
the entire application. In some regards this is useful because the expected
traffic on the deployed application is very low and CGI means that there is no
resource usage at all.

However, if I want to "optimize" (and bear in mind this is a toy project for
me, so premature optimization is one of the goals), then each invocation should
only consume the resources (database connections, filesystem access) required to
process that single path.

The separation is also useful in case I ever wish to split the application into
actual separate scripts, which I am still undecided on, mainly because it makes
deployment more challenging.

## Source layout

| File             | Purpose                                              |
|------------------|------------------------------------------------------|
| `deskd.c`        | Entry point, argument handling, and CGI routing       |
| `cgi.c` / `cgi.h`| CGI response helpers, form/cookie parsing, CSRF, dates|
| `db.c` / `db.h`  | SQLite database access and query functions            |
| `natural.c` / `natural.h` | Natural sort collation for SQLite             |
| `compat.h`       | Portability shims for non-OpenBSD platforms           |
| `about.c`        | GET /about handler                                   |
| `dateform.c`     | GET /book handler (date picker and redirect)         |
| `bookingform.c`  | GET /book/\<date\> handler (desk selection form)     |
| `book.c`         | POST /book/\<date\> handler (create booking)         |
| `bookings.c`     | GET / handler (list user's bookings)                 |
| `cancel.c`       | POST / handler (cancel a booking)                    |
| `html/`          | Static HTML fragments embedded at compile time via `#embed` |
| `sql/`           | SQL statements embedded at compile time via `#embed`  |

## Dependencies

The only external dependency is `sqlite3`, which provides the database library.
Everything else uses C libraries available in OpenBSD base:

- `stdio.h`, `stdlib.h`, `string.h` — standard I/O, memory, and string handling
- `time.h` — date parsing and formatting (`strptime`, `strftime`, `timegm`)
- `ctype.h` — character classification for URL decoding and natural sort
- `arc4random_buf` — CSRF token generation (OpenBSD base, also available on macOS)
- `timingsafe_bcmp` — constant-time comparison for CSRF validation (OpenBSD base,
  also available on macOS)
- `reallocarray` — overflow-checked array allocation (OpenBSD base; shim provided
  for platforms that lack it via `compat.h`)

## Database

I've gone back and forth on using a database or just the filesystem for storing
desk bookings.
The filesystem does not require any additional dependencies, but it is an
absolute pain to set up and maintain, and the performance is not great.
The sqlite database is straightforward and embedded, so no need to deploy a
separate database server.
It is also much more performant, and some logic around uniqueness and sorting
can be offloaded to the database.
This keeps the application code much simpler, which means less to maintain, and
that is a good thing for a toy project.

Yes, I know that sqlite can have locking issues. I consider that the intended
deployment of this application is for at most 30 users with essentially no traffic,
and very little change of concurrent user access, so I don't think it is a problem.
Even if you wanted to scale this out, there are lots of ways that sqlite can be
configured to deal with this; I just haven't had to worry about it yet.

## Testing

There is a set of integration tests in `./test`. These are run using a custom test
runner in `./test/test.pl` and test cases are defined in `.test` case files located
alongside the test runner.

To run the tests, build the application and then run:
```
perl test/test.pl <path to deskd binary>
```

### Case files

An integration test case is a file containing a series of inputs and outputs for
the test runner to read and use for testing.

The format is as follows:
```
Test Case Description
  A single line string describing the test case.
---
Desks Database Seed Data
  Each desk name is a separate line.
---
Bookings Database Seed Data
  Should be in the format of:
    <user>,<desk>,<day>
  The date is in RFC3339 format.
  Each line is a separate booking.
---
Request Environment Variables
  These are used to invoke the application so should emulate CGI environment variables.
  Each variable is a separate line.
  Each line is a key=value pair.
---
Request Body
  The request body is the raw HTTP request body.
  Newline characters are normalised across the case.
---
Expected Bookings Database Data
  Should be in the format of:
    <user>,<desk>,<day>
  The date is in RFC3339 format.
  Each line is a separate booking.
  Bookings should be sorted by user, then desk, then day.
---
Expected Response
  This is the expected output of the application on STDOUT.
  It is interpreted as a regex in the format:
    qr/\A<value>\z/s
  However it will be passed through quotemeta() before being used.
  In order to include regular expressions in the output they should be wrapped in
  the sequence:
    REGEX{<value>}
  The <value> will be exluded from any quoting and the REGEX{} parts will be removed.
  This makes it possible to ignore or forward capture parts of the response such
  as CSRF tokens.
```
