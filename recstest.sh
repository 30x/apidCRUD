#! /bin/bash
#	recstest.sh IDS
# dump all records (by default)
# the API is GET on /db/_table/{table_name} aka getDbRecords

TABLE=bundles
FIELDS=id,name
API_PATH=db/_table
IDS=${1:-\*}
if [[ "$IDS" == \* ]]; then
	IDS=
fi

out=$(./appcurl.sh GET "$API_PATH/$TABLE?ids=$IDS&fields=$FIELDS")
xstat=$?

echo "$out"
# echo "$out" | jq -r -S .Record[].id
exit $xstat
