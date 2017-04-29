package apidCRUD

import (
	"fmt"
	"net/http"
	"encoding/json"
	"io"
)

// apiHandlerRet is the return type from an apiHandler function.
type apiHandlerRet struct {
	code int
	data interface{}
}

// apiHandlerArg is the type of the parameter to an apiHandler function.
type apiHandlerArg struct {
	req *http.Request
}

// apiHandler is the type an API handler function.
type apiHandler func(apiHandlerArg) apiHandlerRet

// type verbMap maps each wired verb for a given path, to its handler function.
type verbMap struct {
	path string
	methods map[string]apiHandler
}

// type apiDesc describes the wiring for one API.
type apiDesc struct {
	path string
	verb string
	handler apiHandler
}

// type apiWiring is the state needed to dispatch an incoming API call.
type apiWiring struct {
	pathsMap map[string]verbMap
}

// newApiWiring returns an API configuration, after adding the APIs
// from the given table.  basePath is prefixed to the paths
// specified in the table items.
func newApiWiring(basePath string, tab []apiDesc) (*apiWiring) {  // nolint
	pm := make(map[string]verbMap)
	apiws := &apiWiring{pm}
	for _, b := range(tab) {
		apiws.AddApi(basePath + b.path, b.verb, b.handler)
	}
	return apiws
}

// AddApi() configures the wiring for one path and verb to their handler.
func (apiws *apiWiring) AddApi(path string,
		verb string,
		handler apiHandler) { // nolint
	vmap, ok := apiws.pathsMap[path]
	if !ok {
		vmap = verbMap{path: path, methods: map[string]apiHandler{}}
		apiws.pathsMap[path] = vmap
	}
	vmap.methods[verb] = handler
}

// GetMaps returns the configured path-to-verbMap mapping,
// for possible range iteration.
func (apiws *apiWiring) GetMaps() map[string]verbMap {
	return apiws.pathsMap
}

// callApiMethod() calls the handler that was configured for
// the given verbMap and verb.
func callApiMethod(vmap verbMap, verb string, req apiHandlerArg) apiHandlerRet {
	verbFunc, ok := vmap.methods[verb]
	if !ok {
		return apiHandlerRet{http.StatusMethodNotAllowed,
			fmt.Errorf(`No handler for %s on %s`, verb, vmap.path)}
	}

	return verbFunc(req)
}

// pathDispatch() is the general handler for all our APIs.
// it is called indirectly thru a closure function that
// supplies the vmap argument.
func pathDispatch(vmap verbMap, w http.ResponseWriter, arg apiHandlerArg) {
	log.Debugf("in pathDispatch: method=%s path=%s",
		arg.req.Method, arg.req.URL.Path)
	defer func() {
		_ = arg.bodyClose()
	}()

	res := callApiMethod(vmap, arg.req.Method, arg)

	rawdata, err := convData(res.data)
	if err != nil {
		writeErrorResponse(w, err)
		return
	}

	w.WriteHeader(res.code)
	_, _ = w.Write(rawdata)

	log.Debugf("in pathDispatch: code=%d", res.code)
}

// convData() converts the interface{} data part returned by
// an apiHandler function.  the return value is a byte slice.
// if the data is json, it essentially gets ascii-fied.
func convData(data interface{}) ([]byte, error) {
	switch data := data.(type) {
	case []byte:
		return data, nil
	case string:
		return []byte(data), nil
	default: // json conversion
		return json.Marshal(data)
	}
}

// writeErrorResponse() writes to the ResponseWriter,
// the given error's message, and logs it.
func writeErrorResponse(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	msg := err.Error()
	data, _ := convData(ErrorResponse{code,msg})

        w.WriteHeader(code)
        _, _ = w.Write(data)

        log.Errorf("error handling API request: %s", msg)
}

// ----- methods for apiHandlerArg

// parseForm() is a wrapper for http.Request.ParseForm().
func (arg *apiHandlerArg) parseForm() error {
	if arg.req.Method == "failmefortestingpurposes" {
		return fmt.Errorf("illegal verb")
	}
	return arg.req.ParseForm()
}

// formValue() is a wrapper for http.Request.FormValue().
func (arg *apiHandlerArg) formValue(name string) string {
	return arg.req.FormValue(name)
}

// getBody() is an accessor for http.Request.Body .
func (arg *apiHandlerArg) getBody() io.ReadCloser {
	return arg.req.Body
}

// bodyClose() is a wrapper for http.Request.Body.Close()
func (arg *apiHandlerArg) bodyClose() error {
	return arg.req.Body.Close()
}
