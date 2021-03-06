#! /bin/bash
#	crtest.sh NRECS
# create the given number of new records.
# the API is POST /db/_table/{table_name} aka createDbRecords .

# output the json for a single record
mkrec()
{
	local i=$1
	cat<<EOF
{"keys":["name", "uri"], "values":["name$i", "host$i"]}
EOF
}

# output json for the given number of records.
mk_nrecs()
{
	local n=$1
	local sep=""
	local i
	((i=1))
	while ((i <= n)); do
		echo -n "$sep"
		mkrec $i
		sep=","
		((i++))
	done
}

# ----- start of mainline code
PROGDIR=$(cd "$(dirname "$0")" && /bin/pwd)
. "$PROGDIR/tester-env.sh" || exit 1
. "$PROGDIR/test-common.sh" || exit 1

NRECS=${1:-2}
RESOURCES="[$(mk_nrecs "$NRECS")]"
BODY="{\"records\":$RESOURCES}"
# echo 1>&2 "BODY=$BODY"

out=$(apicurl POST "db/_table/$TABLE_NAME" -v -d "$BODY")
echo "$out"

# echo "$out" | jq -S '.ids[]'
