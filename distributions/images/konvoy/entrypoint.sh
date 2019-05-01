#!/usr/bin/env sh
set -e

# if the first argument look like a parameter (i.e. start with '-'), run Konvoy
if [ "${1#-}" != "$1" ]; then
	set -- konvoy "$@"
fi

if [ "$1" = 'konvoy' ]; then
	# set the log level if the $loglevel variable is set
	if [ -n "$loglevel" ]; then
		set -- "$@" --log-level "$loglevel"
	fi
fi

exec "$@"
