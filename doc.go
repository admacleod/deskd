// Deskd is attempted to be laid out in a sensible sort of fashion.
// What I'm going for is to have each path as individual as possible,
// so rather than creating objects to handle all the required
// functionality, instead each path should be able to be considered
// as a totally separate script or application.
//
// The rationale behind this separation is that this software is
// intended to run as a CGI script; this means that every request
// results in a new invocation of the entire application.
// In some regards this is useful because the expected traffic on
// the deployed application is very low and CGI means that there is
// no resource usage at all. However, if I want to "optimize" (and
// bear in mind this is a toy project for me, so premature optimization
// is one of the goals), then each invocation should only consume the
// resources (database connections, filesystem access) required
// to process that single path.
// The separation is also useful in case I ever wish to split
// the application into actual separate scripts, which I am still
// undecided on, mainly because it makes deployment more challenging.

package main
