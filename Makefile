CC ?= cc
CFLAGS += -Wall -Wextra -pedantic -std=c23 -Werror -I/usr/local/include
LDFLAGS += -L/usr/local/lib -lsqlite3

SRCS = deskd.c cgi.c db.c natural.c about.c dateform.c bookings.c \
       bookingform.c book.c cancel.c
OBJS = ${SRCS:.c=.o}

deskd: ${OBJS}
	${CC} -o $@ ${OBJS} ${LDFLAGS}

.c.o:
	${CC} ${CFLAGS} -c $<

clean:
	rm -f ${OBJS} deskd

.PHONY: clean
