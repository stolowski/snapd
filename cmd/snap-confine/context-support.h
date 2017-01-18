/*
 * Copyright (C) 2017 Canonical Ltd
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License version 3 as
 * published by the Free Software Foundation.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 *
 */

#ifndef SNAP_CONFINE_CONTEXT_SUPPORT_H
#define SNAP_CONFINE_CONTEXT_SUPPORT_H

#include "error.h"

/**
* Error domain for errors related to snap context handling.
**/
#define SC_CONTEXT_DOMAIN "context"

char *sc_nonfatal_context_get_from_snapd(const char *snap_name,
					 struct sc_error **errorp);
void sc_context_set_environment(const char *context);

#endif
