#!/usr/bin/env bash

modules()
{
	local pref="github\.com/30x/apidCRUD"
	go list ./... \
	| egrep -v "^$pref/vendor|^$pref/obs|^$pref/cmd"
}


# ----- start of mainline code

cdir=cov
ctxt=$cdir/coverage.txt
/bin/rm -f "$ctxt"

mkdir -p "$cdir"

set -e
# echo "mode: atomic" > "$ctxt"

/bin/rm -f "$prof"

for d in $(modules); do
    prof=$cdir/profile.out
    go test "-coverprofile=$prof" -covermode=atomic $d
    if [ -f "$prof" ]; then
        head -2 "$prof" >> "$ctxt"
        rm "$prof"
    fi
done
go tool cover "-html=$ctxt" -o "$cdir/cover.html"
