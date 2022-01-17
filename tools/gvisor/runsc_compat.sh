#!/bin/bash

BIN="/usr/bin/runsc"

IS_EXEC=0
for arg; do
    if [ "$arg" == "exec" ]; then
        IS_EXEC=1
    fi
done


# GVisor's runsc does not support "-d" which is passed to it from conmon.
# runc supports "-d" for running detached so translate the "-d" argument to
# the "-detach" flavor supported by runsc.
if [ $IS_EXEC -eq 1 ]; then
    declare -a NEWARGS
    for arg; do
        if [ "$arg" == "-d" ]; then
            NEWARGS+=("-detach")
        else
            NEWARGS+=("$arg")
        fi
    done
    set -- "${NEWARGS[@]}"
fi

exec "$BIN" "$@"