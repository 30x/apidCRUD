package apidCRUD

import (
	"testing"
	"fmt"
	"strings"
	"os"
	"net/http"
	"database/sql"
	"github.com/30x/apid-core"
	"github.com/30x/apid-core/factory"
)

var testServices = factory.DefaultServicesFactory()

// TestMain() is called by the test framework before running the tests.
// we use it to initialize the log variable.
func TestMain(m *testing.M) {
	// do this in case functions under test need to log something
	apid.Initialize(testServices)
	log = apid.Log()

	// required boilerplate
	os.Exit(m.Run())
}

// mySplit() is like strings.Split() except that
// it returns a 0-length slice when s is the empty string.
func mySplit(str string, sep string) []string {
	if str == "" {
		return []string{}
	}
	return strings.Split(str, sep)
}

// ----- unit tests for mkSQLRow()

func mkSQLRow_Checker(t *testing.T, i int, N int) {
	fn := "mkSQLRow"
	res := mkSQLRow(N)
	if len(res) != N {
		t.Errorf("#%d: %s(%d) failed", i, fn, N)
		return
	}
	for _, v := range res {
		_, ok := v.(*sql.RawBytes)
		if !ok {
			t.Errorf("#%d: %s(%d) sql conversion error", i, fn, N)
			return
		}
	}
}

func Test_mkSQLRow(t *testing.T) {
	M := 5
	for i := 0; i < M; i++ {
		mkSQLRow_Checker(t, i, i)
	}
}

// ----- unit tests for notImplemented()

func Test_notImplemented(t *testing.T) {
	fn := "notImplemented"
	xcode := http.StatusNotImplemented
	res := notImplemented()
	if res.code != xcode {
		t.Errorf("%s returned code %d; expected %d",
			fn, res.code, xcode)
	}
	if res.data == nil {
		t.Errorf("%s returned nil error; expected non-nil", fn)
	}
}

// ----- unit tests for validateSQLValues()

func genList(form string, N int) []string {
	ret := make([]string, N)
	for i := 0; i < N; i++ {
		ret[i] = fmt.Sprintf(form, i)
	}
	return ret
}

func genListInterface(form string, N int) []interface{} {
	ret := make([]interface{}, N)
	for i := 0; i < N; i++ {
		ret[i] = fmt.Sprintf(form, i)
	}
	return ret
}

func sqlValues_Checker(t *testing.T, form string, N int) {
	fn := "validateSQLValues"
	values := genListInterface(form, N)
	err := validateSQLValues(values)
	if err != nil {
		t.Errorf("%s(...) failed on length=%d", fn, N)
	}
}

func Test_validateSQLValues(t *testing.T) {
	M := 5
	for j := 0; j < M; j++ {
		sqlValues_Checker(t, "V%d", j)
	}

	// empty values OK
	sqlValues_Checker(t, "", 3)
}

// ----- unit tests for validateSQLKeys()

func sqlKeys_Checker(t *testing.T, form string, N int, xsucc bool) {
	fn := "validateSQLKeys"
	values := genList(form, N)
	err := validateSQLKeys(values)
	if xsucc != (err == nil) {
		msg := "true"
		if err != nil {
			msg = err.Error()
		}
		t.Errorf(`%s("%s"...)=%s; expected %t`,
			fn, form, msg, xsucc)
	}
}

func Test_validateSQLKeys(t *testing.T) {
	M := 3
	for j := 0; j < M; j++ {
		sqlKeys_Checker(t, "K%d", j, true)
	}

	// numeric key not OK
	sqlKeys_Checker(t, "%d", 1, false)

	// empty key not OK
	sqlKeys_Checker(t, "", 1, false)
}

// ----- unit tests for nstring()

func nstring_Checker(t *testing.T, s string, n int) {
	fn := "nstring"
	res := nstring(s, n)
	rlist := strings.Split(res, ",")
	if n == 0 {
		// this must be handled as a special case
		// because strings.Split() returns a list of length 1
		// on empty string.
		if res != "" {
			t.Errorf(`%s("%s",%d)="%s"; expected ""`,
				fn, s, n, res)
		}
		return
	} else if n != len(rlist) {
		t.Errorf(`%s("%s",%d)="%s" failed split test`,
			fn, s, n, res)
		return
	}
	for _, v := range rlist {
		if v != s {
			t.Errorf(`%s("%s",%d) bad item "%s"`,
				fn, s, n, v)
		}
	}
}

func Test_nstring(t *testing.T) {
	M := 3
	for j := 0; j < M; j++ {
		nstring_Checker(t, "", j)
		nstring_Checker(t, "abc", j)
	}
}

// ----- unit tests for errorRet()

type errorRet_TC struct {
	code int
	msg string
}

var errorRet_Tab = []errorRet_TC {
	{ 1, "abc" },
	{ 2, "" },
	{ 3, "xyz" },
}

func errorRet_Checker(t *testing.T, i int, code int, msg string) {
	fn := "errorRet"
	err := fmt.Errorf("%s", msg)
	res := errorRet(code, err)
	if code != res.code {
		t.Errorf(`#%d: %s returned (%d,); expected %d`,
			i, fn, res.code, code)
		return
	}
	eresp, ok := res.data.(ErrorResponse)
	if !ok {
		t.Errorf(`#%d: %s ErrorResponse conversion error`, i, fn)
		return
	}
	if code != eresp.Code {
		t.Errorf(`#%d: %s ErrorResponse.Code=%d; expected %d`,
			i, fn, eresp.Code, code)
		return
	}
	if msg != eresp.Message {
		t.Errorf(`#%d: %s ErrorResponse.Message="%s"; expected "%s"`,
			i, fn, eresp.Message, msg)
	}
}

func Test_errorRet(t *testing.T) {
	for i, tc := range errorRet_Tab {
		errorRet_Checker(t, i, tc.code, tc.msg)
	}
}

// ----- unit tests for mkIdClause()

func fakeParams(paramstr string) map[string]string {
	ret := map[string]string{}
	if paramstr == "" {
		return ret
	}
	var name, value string
	strlist := strings.Split(paramstr, "&")
	for _, s := range strlist {
		if s == "" {
			continue
		}
		words := strings.SplitN(s, "=", 2)
		switch len(words) {
		case 1:
			name = words[0]
			value = ""
		case 2:
			name = words[0]
			value = words[1]
		default:
			name = ""
			value = ""
		}
		// fmt.Printf("in fakeParams, name=%s, value=%s\n", name, value)
		ret[name] = value
	}
	return ret
}

type idclause_TC struct {
	paramstr string
	xres string
	xids string
	xsucc bool
}

var idclause_Tab = []idclause_TC {
	{ "id_field=id&id=123", "WHERE id = ?", "123", true },
	{ "id_field=id&ids=123", "WHERE id in (?)", "123", true },
	{ "id_field=id&ids=123,456", "WHERE id in (?,?)", "123,456", true },
	{ "id_field=id", "", "", true },
}

func aToIdList(s string) []interface{} {	// nolint
	if s == "" {
		return []interface{}{}
	}
	slist := strings.Split(s, ",")
	N := len(slist)
	ret := make([]interface{}, len(slist))
	for i := 0; i < N; i++ {
		ret[i] = aToIdType(slist[i])
	}
	return ret
}

// convert idlist (list of int64) to ascii csv.
func idListToA(idlist []interface{}) (string, error) {
	alist := make([]string, len(idlist))
	for i, ival := range idlist {
		val, ok := ival.(int64)
		if !ok {
			return "",
				fmt.Errorf(`idListToA conversion error on "%s"`, ival)
		}
		alist[i] = idTypeToA(val)
	}
	return strings.Join(alist, ","), nil
}

func mkIdClause_Checker(t *testing.T, i int, tc idclause_TC) {
	fn := "mkIdClause"
	params := fakeParams(tc.paramstr)
	res, idlist, err := mkIdClause(params)
	if tc.xsucc != (err == nil) {
		msg := errRep(err)
		t.Errorf(`#%d: %s([%s]) returned status=[%s]; expected [%t]`,
			i, fn, tc.paramstr, msg, tc.xsucc)
		return
	}
	if err != nil {
		return
	}
	if tc.xres != res {
		t.Errorf(`#%d: %s([%s]) returned "%s"; expected "%s"`,
			i, fn, tc.paramstr, res, tc.xres)
	}

	resids, err := idListToA(idlist)
	if err != nil {
		t.Errorf(`#%d: %s idListToA error "%s"`, i, fn, err)
	}
	if tc.xids != resids {
		t.Errorf(`#%d: %s([%s]) idlist=[%s]; expected [%s]`,
			i, fn, tc.paramstr, resids, tc.xids)
	}
}

func Test_mkIdClause(t *testing.T) {
	for i, tc := range idclause_Tab {
		mkIdClause_Checker(t, i, tc)
	}
}

// ----- unit tests for mkIdClauseUpdate()

var mkIdClauseUpdate_Tab = []idclause_TC {
	{ "id_field=id&id=123", "WHERE id = 123", "", true },
	{ "id_field=id&ids=123", "WHERE id in (123)", "", true },
	{ "id_field=id&ids=123,456", "WHERE id in (123,456)", "", true },
	{ "id_field=id", "", "", true },
}

func mkIdClauseUpdate_Checker(t *testing.T, i int, tc idclause_TC) {
	fn := "mkIdClauseUpdate"
	params := fakeParams(tc.paramstr)
	res, err := mkIdClauseUpdate(params)
	if tc.xsucc != (err == nil) {
		msg := errRep(err)
		t.Errorf(`#%d: %s([%s]) returned status=[%s]; expected [%t]`,
			i, fn, tc.paramstr, msg, tc.xsucc)
		return
	}
	if err != nil {
		return
	}
	if tc.xres != res {
		t.Errorf(`#%d: %s([%s]) returned "%s"; expected "%s"`,
			i, fn, tc.paramstr, res, tc.xres)
	}
}

func Test_mkIdClauseUpdate(t *testing.T) {
	for i, tc := range mkIdClauseUpdate_Tab {
		mkIdClauseUpdate_Checker(t, i, tc)
	}
}

// ----- unit tests for idTypesToInterface()

type idTypesToInterface_TC struct {
	ids string
}

var idTypesToInterface_Tab = []string {
	"",
	"987",
	"654,321",
	"987,654,32,1,0",
}

func idTypesToInterface_Checker(t *testing.T, i int, tc string) {
	fn := "idTypesToInterface"
	alist := strings.Split(tc, ",")
	if tc == "" {
		alist = []string{}
	}
	res := idTypesToInterface(alist)
	str, err := idListToA(res)
	if err != nil {
		t.Errorf(`#%d: %s idListToA error "%s"`, i, fn, err)
	}
	if str != tc {
		t.Errorf(`#%d: %s("%s") = "%s"; expected "%s"`,
			i, fn, tc, str, tc)
		return
	}
}

func Test_idTypesToInterface(t *testing.T) {
	for i, tc := range idTypesToInterface_Tab {
		idTypesToInterface_Checker(t, i, tc)
	}
}


// ----- unit tests for mkSelectString()

type mkSelectString_TC struct {
	paramstr string
	xres string
	xids string
	xsucc bool
}

var mkSelectString_Tab = []mkSelectString_TC {
	{"table_name=T&id_field=id&id=456&fields=a&limit=1&offset=0",
		"SELECT a FROM T WHERE id = ? LIMIT 1 OFFSET 0",
		"456", true},
	{"table_name=T&id_field=id&ids=123,456&fields=a,b,c&limit=1&offset=0",
		"SELECT a,b,c FROM T WHERE id in (?,?) LIMIT 1 OFFSET 0",
		"123,456", true},
}

// run one tc case
func mkSelectString_Checker(t *testing.T, i int, tc mkSelectString_TC) {
	fn := "mkSelectString"
	params := fakeParams(tc.paramstr)
	// fmt.Printf("in %s_Checker, params=%s\n", fn, params)
	res, idlist, err := mkSelectString(params)
	if tc.xsucc != (err == nil) {
		msg := errRep(err)
		t.Errorf(`#%d: %s returned status [%s]; expected [%t]`,
			i, fn, msg, tc.xsucc)
		return
	}
	if err != nil {
		return
	}
	if tc.xres != res {
		t.Errorf(`#%d: %s returned "%s"; expected "%s"`,
			i, fn, res, tc.xres)
		return
	}
	ids, err := idListToA(idlist)
	if err != nil {
		t.Errorf(`#%d: %s idListToA error "%s"`, i, fn, err)
	}
	if tc.xids != ids {
		t.Errorf(`#%d: %s returned ids "%s"; expected "%s"`,
			i, fn, ids, tc.xids)
	}
}

func Test_mkSelectString(t *testing.T) {
	for i, tc := range mkSelectString_Tab {
		mkSelectString_Checker(t, i, tc)
	}
}

// ----- unit tests for getBodyRecord()

type getBodyRecord_TC struct {
	data string
	keys string
	values string
}

// turn a list of strings masked as interfaces, back to list of strings.
func unmaskStrings(ilist []interface{}) []string {
	N := len(ilist)
	ret := make([]string, N)
	for i := 0; i < N; i++ {
		s, _ := ilist[i].(string)
		ret[i] = s
	}
	return ret
}

// turn a list of strings, into strings masked by interface.
func maskStrings(slist []string) []interface{} {
	N := len(slist)
	ret := make([]interface{}, N)
	for i := 0; i < N; i++ {
		ret[i] = slist[i]
	}
	return ret
}

var getBodyRecord_Tab = []getBodyRecord_TC {
	{`{"Records":[{"Keys":[], "Values":[]}]}`,
		"",
		""},
	{`{"Records":[{"Keys":["k1","k2","k3"], "Values":["v1","v2","v3"]}]}`,
		"k1,k2,k3",
		"v1,v2,v3"},
	{`{"Records":[{"Keys":["k1","k2","k3"], "Values":["v1","v2","v3"]},{"Keys":["k4","k5","k6"], "Values":["v4","v5","v6"]}]}`,
		"k1,k2,k3&k4,k5,k6",
		"v1,v2,v3&v4,v5,v6"},
}

func getBodyRecord_Checker(t *testing.T, testno int, tc getBodyRecord_TC) {
	fn := "getBodyRecord"

	rdr := strings.NewReader(tc.data)
	req, _ := http.NewRequest(http.MethodPost, "/xyz", rdr)

	tckeys := strings.Split(tc.keys, "&")
	tcvalues := strings.Split(tc.values, "&")
	nkeys := len(tckeys)

	body, err := getBodyRecord(apiHandlerArg{req})
	if err != nil {
		t.Errorf("#%d: %s([%s]) failed, error=%s",
			testno, fn, tc.data, err)
	}
	records := body.Records
	nrecs := len(records)

	if nkeys != nrecs {
		t.Errorf(`#%d: %s returned Records length=%d; expected %d`,
			testno, fn, nrecs, nkeys)
	}
	for j := 0; j < nrecs; j++ {
		rec := records[j]
		keystr := strings.Join(rec.Keys, ",")
		if tckeys[j] != keystr {
			t.Errorf(`#%d %s Record[%d] keys=%s; expected %s`,
				testno, fn, j, keystr, tckeys[j])
		}
		valstr := strings.Join(unmaskStrings(rec.Values), ",")
		if tcvalues[j] != valstr {
			t.Errorf(`#%d %s Record[%d] values=%s; expected %s`,
				testno, fn, j, valstr, tcvalues[j])
		}
	}
}

func Test_getBodyRecord(t *testing.T) {
	for testno, tc := range getBodyRecord_Tab {
		getBodyRecord_Checker(t, testno, tc)
	}
}

// ----- unit tests for convTableNames() and grabNameField()

type convTableNames_TC struct {
	names string
}

var convTableNames_Tab = []convTableNames_TC {
	{""},
	{"a"},
	{"a,b"},
	{"abc,def,ghi"},
}

// mimicTableNamesQuery() returns an object that mimics the return from
// the query to the "tables" table.
func mimicTableNamesQuery(names []string) []*KVRecord {
	N := len(names)
	ret := make([]*KVRecord, N)
	for i := 0; i < N; i++ {
		Keys := []string{"name"}
		Values := []interface{}{names[i]}
		ret[i] = &KVRecord{Keys, Values}
	}
	return ret
}

func convTableNames_Checker(t *testing.T, testno int, tc convTableNames_TC) {
	fn := "convTableNames"
	names := mySplit(tc.names, ",")
	obj := mimicTableNamesQuery(names)
	// fmt.Printf("obj=%s\n", obj)
	res, err := convTableNames(obj)
	if err != nil {
		t.Errorf("#%d: %s([%s]) returned error", testno, fn, tc.names)
		return
	}
	resJoin := strings.Join(res, ",")
	if tc.names != resJoin {
		t.Errorf(`#%d: %s([%s]) = "%s"; expected "%s"`,
			testno, fn, tc.names, resJoin, tc.names)
	}
}

func Test_convTableNames(t *testing.T) {
	for testno, tc := range convTableNames_Tab {
		convTableNames_Checker(t, testno, tc)
	}
}

// ----- unit tests for validateRecords()

type validateRecords_TC struct {
	desc string
	xsucc bool
}

// mkRecords() turns a description string into an array of records.
// the description is parsed as follows.
// record descriptions are separated by ';' chars.
// within each record, '|' separates the list of keys from the list of values.
// the keys (names) are comma-separated.
func mkRecords(desc string) []KVRecord {
	desclist := mySplit(desc, ";")
	nrecs := len(desclist)
	ret := make([]KVRecord, nrecs)
	for i, rdesc := range desclist {
		parts := mySplit(rdesc, "|")
		ret[i].Keys = mySplit(parts[0], ",")
		ret[i].Values = maskStrings(mySplit(parts[1], ","))
	}
	return ret
}

var validateRecords_Tab = []validateRecords_TC {
	{"", true},		// 0 records
	{"k1,k2,k3|v1,v2,v3", true},  // 1 record, valid
	{"k1,,k3|v1,v2,v3", false},   // 1 record, invalid
	{"k1,k2,k3|v1,v2,v3;k4|v4", true}, // 2 records, valid
	{"k1,k2,k3|v1,v2,v3;|v4", false}, // 2 records, invalid
}

func validateRecords_Checker(t *testing.T, testno int, tc validateRecords_TC) {
	fn := "validateRecords"
	records := mkRecords(tc.desc)
	res := validateRecords(records)
	if tc.xsucc != (res == nil) {
		t.Errorf(`#%d: %s([%s]) = [%s]; expected %t`,
			testno, fn, tc.desc, errRep(nil), tc.xsucc)
	}
}

func Test_validateRecords(t *testing.T) {
	for testno, tc := range validateRecords_Tab {
		validateRecords_Checker(t, testno, tc)
	}
}

// ----- unit tests for convValues()

// inputs and outputs for one convValues testcase.
type convValues_TC struct {
	arg string
}

// table of convValues testcases.
var convValues_Tab = []convValues_TC {
	{ "" },
	{ "abc" },
	{ "abc,def" },
	{ "abc,def,ghi" },
}

func strToSQLValues(arg string) []interface{} {
	args := mySplit(arg, ",")
	N := len(args)
	ret := make([]interface{}, N)
	for i, s := range args {
		rb := sql.RawBytes(s)
		ret[i] = &rb;
	}
	return ret
}

// return something that can't be converted by convValues().
func mkIllegalValues() []interface{} {
	ret := make([]interface{}, 1)
	val := 555
	ret[0] = &val
	return ret
}

// run one testcase for function convValues.
func convValues_Checker(t *testing.T, testno int, tc convValues_TC) {
	fn := "convValues"
	argInter := strToSQLValues(tc.arg)
	err := convValues(argInter)
	if err != nil {
		t.Errorf(`#%d: %s([%s]) failed [%s]`,
			testno, fn, tc.arg, err)
	}
	argStrings := unmaskStrings(argInter)
	resultStr := strings.Join(argStrings, ",")
	if tc.arg != resultStr {
		t.Errorf(`#%d: %s("%s")="%s"; expected "%s"`,
			testno, fn, tc.arg, resultStr, tc.arg)
	}
}

// main test suite for convValues().
func Test_convValues(t *testing.T) {
	for testno, tc := range convValues_Tab {
		convValues_Checker(t, testno, tc)
	}
}

// test suite for testing error return.
func Test_convValues_illegal(t *testing.T) {
	fn := "convValues"
	vals := mkIllegalValues()
	err := convValues(vals)
	if err == nil {
		t.Errorf(`%s on illegal value failed to return error`, fn)
	}
}
