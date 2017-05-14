#! /bin/bash
#	tester.sh
# try out a variety of APIs, and do some cursory tests.
# this assumes the server is already running.

get_nrecs()
{
	"$TESTS_DIR/recstest.sh" '*' 2>/dev/null \
	| jq -S '.Records[].Values[0]' | grep -c ""
}

get_rec_uri()
{
	local ID=$1
	local TABLE=bundles
	local FIELDS=id,uri
	local API_PATH=db/_table
	apicurl GET "$API_PATH/$TABLE/$ID?fields=$FIELDS" 2>/dev/null \
	| jq -r -S '.Records[].Values[1]'
}

list_tables()
{
	echo ".tables" \
	| sqlite3 "$DBFILE" 2>/dev/null \
	| tr ' ' '\n' | grep -v '^$'
}

TestHeader()
{
	echo -n "# $* - "
}

AssertOK()
{
	if [ $? -ne 0 ]; then
		echo "FAIL - $*"
		exit 1
	else
		echo OK
	fi
}

# ----- start of mainline
TESTS_DIR=functests

. "$TESTS_DIR/tester-env.sh" || exit 1
. "$TESTS_DIR/test-common.sh" || exit 1

# start clean
TestHeader creating empty database
"$TESTS_DIR/mkdb.sh"
AssertOK "database initialization"

TestHeader checking _tables_ "(tabtest.sh)"
out=$("$TESTS_DIR/tabtest.sh" 2>/dev/null | sort | tr '\n' ' ')
tabs=( $out )
exp=( bundles file nothing users )
[[ "${tabs[*]}" == "${exp[*]}" ]]
AssertOK "tabtest.sh expected [${exp[*]}], got [${tabs[*]}]"

TestHeader "adding a few records (crtest.sh)"
nrecs=7
out=$("$TESTS_DIR/crtest.sh" "$nrecs" 2>/dev/null | jq -S '.Ids[]')
nc=$(echo "$out" | grep -c "")
[[ "$nc" == "$nrecs" ]]
AssertOK "crtest.sh expected $nrecs, got $nc"

TestHeader "read one record (rectest.sh)"
out=$("$TESTS_DIR/rectest.sh" 7 2>/dev/null)
[[ "$out" == 7 ]]
AssertOK "rectest.sh expected 7, got $out"

TestHeader "reading the records (recstest.sh)"
total=$(get_nrecs)
[[ "$total" == "$nc" ]]
AssertOK "recstest.sh expected $total, got $nc"

TestHeader "deleting a record (deltest.sh)"
nc=$("$TESTS_DIR/deltest.sh" 7 2>/dev/null)
[[ "$nc" == 1 ]]
AssertOK "deltest.sh expected 1, got $nc"

TestHeader "checking total number of records (recstest.sh)"
total=$(get_nrecs)
((xtotal=nrecs-1))
[[ "$total" == "$xtotal" ]]
AssertOK "deltest.sh expected $xtotal, got $total"

TestHeader "deleting more records (delstest.sh)"
nc=$("$TESTS_DIR/delstest.sh" 2,3,4 2>/dev/null)
[[ "$nc" == 3 ]]
AssertOK "delstest.sh expected 3, got $nc"

TestHeader "updating a record (uptest.sh)"
nc=$("$TESTS_DIR/uptest.sh" 5 2>/dev/null)
[[ "$nc" == 1 ]]
AssertOK "uptest.sh expected 1, got $nc"

TestHeader "check rec 6 uri before update (get_rec_uri)"
uri1=$(get_rec_uri 6)
[[ $uri1 != "" ]]
AssertOK "uri1 empty"

TestHeader "updating 2 records (upstest.sh)"
nc=$("$TESTS_DIR/upstest.sh" 1,6 2>/dev/null)
[[ "$nc" == 2 ]]
AssertOK "upstest.sh expected 2, got $nc"

TestHeader "checking the update (get_rec_uri)"
uri2=$(get_rec_uri 6)
[[ "$uri1" != "$uri2" ]]
AssertOK "update did not change uri = $uri1"

TestHeader "try writing a small file and reading it back (rwftest.sh)"
"$TESTS_DIR/rwftest.sh" cmd/apidCRUD/main.go > /dev/null 2>&1
AssertOK file comparison

TestHeader "trying tables creation (crtabtest.sh)"
out=$("$TESTS_DIR/crtabtest.sh" X Y Z 2>/dev/null)
out=$(list_tables | grep -c '^[XYZ]$')
[[ "$out" == 3 ]]
AssertOK "tables creation"

TestHeader "trying table deletion (deltabtest.sh)"
out=$("$TESTS_DIR/deltabtest.sh" X Y Z 2>/dev/null)
out=$(list_tables | grep '^$[XYZ]$')
[[ $? != 0 ]]  # the grep should have failed
AssertOK "table deletion"

TestHeader "trying table schema (desctabtest.sh)"
out=$("$TESTS_DIR/desctabtest.sh" users 2>/dev/null)
[[ "$out" != "" ]]
AssertOK "table description"

echo "# all passed"
exit 0
