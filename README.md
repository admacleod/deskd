
[SPDX-FileCopyrightText: 2022 Alisdair MacLeod <copying@alisdairmacleod.co.uk>]::
[SPDX-License-Identifier: AGPL-3.0-only]::

# Desk Booking

This is a fairly simple desk booking system. 
It is purposefully simple to make it easier to deploy.

The code is made available under the [AGPLv3](https://www.gnu.org/licenses/agpl-3.0.en.html).
For commercial licences please get in touch.

`deskd` runs as a daemon serving fcgi over a unix socket.

## Building

Everything is Go, so you can just use `go build` or `go install` to build binaries.
Alternatively there is a `Makefile` that contains prepared commands for building production ready binaries.

## Deployment

`deskd` is very simple and can be run from anywhere and integrated with a webserver like OpenBSD httpd.
It is recommended to run it in a chroot/jail/container of some form to ensure that it cannot be used to compromise a wider system.
`deskd` uses HTTP Basic Authentication so make sure it is served over TLS.

On first start `deskd` will create a sqlite database at whichever location is configured,
and will configure that database with the required schema for `deskd` to operate.
All that is then required is to add records to the database that define the desks and users.

Optionally you can have `deskd` serve a favicon and/or a desk layout map by putting the files `favicon.ico` and `floorplan.png` into the static directory.

## Configuration

`deskd` reads its configuration from the same places (in order of precedence):

- commandline flags passed directly to each tool
- environment variables, prefixed with `DESKD_`

The configuration options are as follows:

| Flag      | Env                | Type      | Default                       | Description                                              |
|-----------|--------------------|-----------|-------------------------------|----------------------------------------------------------|
| `-db`     | `DESKD_DB`         | `string`  | `"test.db"`                   | Location of the database                                 |
| `-socket` | `DESKD_SOCKET`     | `string`  | `"/var/www/run/deskd.socket"` | Location of the socket serving FCGI                      |
| `-static` | `DESKD_STATIC_DIR` | `string`  | `"static"`                    | Path to directory that static assets will be served from |
