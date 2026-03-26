/*
 * Copyright 2026 Alisdair MacLeod <copying@alisdairmacleod.co.uk>
 *
 * This file is part of deskd.
 *
 * deskd is free software: you can redistribute it and/or modify it under the
 * terms of the GNU Affero General Public License as published by the Free
 * Software Foundation, either version 3 of the License, or (at your option)
 * any later version.
 *
 * deskd is distributed in the hope that it will be useful, but WITHOUT ANY
 * WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS
 * FOR A PARTICULAR PURPOSE. See the GNU Affero General Public License for
 * more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with deskd. If not, see <https://www.gnu.org/licenses/>.
 */

#ifndef DESKD_COMPAT_H
#define DESKD_COMPAT_H

/*
 * Compatibility shims for functions available in OpenBSD base
 * but missing on other platforms used for development and testing.
 */

#include <stdint.h>
#include <stdlib.h>
#include <errno.h>

#ifdef __OpenBSD__
#include <unistd.h>	/* pledge, unveil */
#endif

/*
 * reallocarray - safe realloc with overflow checking.
 * Available in OpenBSD and glibc >= 2.26; provide a shim elsewhere
 * (including macOS/Apple libc which does not provide it).
 */
#if !defined(__OpenBSD__) && !defined(__GLIBC__)
static inline void *
reallocarray(void *ptr, const size_t nmemb, const size_t size)
{
	if (size != 0 && nmemb > SIZE_MAX / size) {
		errno = ENOMEM;
		return NULL;
	}
	return realloc(ptr, nmemb * size);
}
#endif

/*
 * pledge and unveil are OpenBSD-specific security mechanisms.
 * On other platforms they are no-ops.
 */
#ifndef __OpenBSD__
static inline int
pledge(const char *promises, const char *execpromises)
{
	(void)promises;
	(void)execpromises;
	return 0;
}

static inline int
unveil(const char *path, const char *permissions)
{
	(void)path;
	(void)permissions;
	return 0;
}
#endif /* !__OpenBSD__ */

#endif /* DESKD_COMPAT_H */
