# Desk Booking

This is a fairly simple desk booking system. 
It is purposefully simple to make it easier to deploy.

The code is made available under the [AGPLv3](https://www.gnu.org/licenses/agpl-3.0.en.html).
For commercial licences please get in touch.

`deskd` is intended to be run as a cgi program on a web server.

## Building

Everything is Go, so you can just use `go build` or `go install` to build binaries.

It is now possible (and strongly recommended) to statically build like so:
```
CGO_ENABLED=1 go build -ldflags="-s -extldflags=-static" -o deskd
```
so that no libraries need to be copied around if running in chroot or scratch containers.

## Configuration

`deskd` reads its configuration from the environment only.

The configuration options are as follows:

| Env        | Type     | Default                            | Description                                                |
|------------|----------|------------------------------------|------------------------------------------------------------|
| `DESKD_DB` | `string` | `"file:/db/deskd.db?cache=shared"` | The DSN used to access a sqlite database storing bookings. |

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

`deskd` is attempted to be laid out in a sensible sort of fashion.
What I'm going for is to have each path as individual as possible.
So rather than creating objects to handle all the required
functionality, instead each path should be able to be considered
as a totally separate script or application.

The rationale behind this separation is that this software is
intended to run as a CGI script; this means that every request
results in a new invocation of the entire application.
In some regards this is useful because the expected traffic on
the deployed application is very low and CGI means that there is
no resource usage at all.
However, if I want to "optimize" (and bear in mind this is a toy
project for me, so premature optimization is one of the goals),
then each invocation should only consume the resources (database
connections, filesystem access) required to process that single
path.
The separation is also useful in case I ever wish to split the
application into actual separate scripts, which I am still 
undecided on, mainly because it makes deployment more challenging.

## Dependencies

External dependencies are kept to a minimum:
- A sqlite driver (no such driver exists in the standard library for good reasons).
  The mattn/go-sqlite3 driver has been selected as it is the most popular and is,
  at the time of writing, the only one in the list at https://go.dev/wiki/SQLDrivers
  and is included in, and passes, the https://github.com/bradfitz/go-sql-test
  test suite. (Yes, I know this is a sort of meaningless bar to cross, but it is at
  least _something_).
- A natural sorting library. This is helpful to handle human-friendly sorting of
  desks, and no such sorting function exists in the standard library.
  The maruel/natural library has been selected as it is both popular, recently
  updated (it slots neatly into the new `slices.Sort` function), has a comprehensive
  test suite, and takes care to reduce memory allocations, which isn't that
  important for this application, but is a useful metric that the author has taken
  some care (or just loves optimization)!

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
---
Request Environment Variables
  These are used to invoke the application so should emulate CGI environment variables.
  Each variable is a separate line.
  Each line is a key=value pair.
---
Expected Bookings Database Data
---
Expected Response
  This is the expected output of the application on STDOUT.
  Headers are not sorted or canonicalized (yeah, not great but whatever).
  Newline characters will be normalised across the case and response.
```
