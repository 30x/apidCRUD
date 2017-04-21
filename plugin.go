package apidCRUD

import (
	"net/http"
	"encoding/json"
	"github.com/30x/apid-core"
)

var (
	log apid.LogService
)

// initPlugin() is called by the apid-core startup
func initPlugin(services apid.Services) (apid.PluginData, error) {
	log = services.Log().ForModule(pluginData.Name)
	log.Printf("in initPlugin")

	initConfig()

	initDB()

	registerHandlers(services.API())

	return pluginData, nil
}

// registerHandlers() register all our handlers with the given service.
func registerHandlers(service apid.APIService) {
	maps := initWiring().getMaps()
	for path, methods := range maps {
		addHandler(service, path, methods)
	}
}

// addHandler() registers the given path with the given service,
// so that it will be handled indirectly by dispatch().
// when an API call is made on this path, the methods argument from
// this context will be suppllied, along with the w and r arguments
// passed in by the service framework.
func addHandler(service apid.APIService, path string, methods verbMap) {
	service.HandleFunc(path,
		func(w http.ResponseWriter, r *http.Request) {
			dispatch(methods, w, r)
		})
}

// dispatch() is the general handler for all our APIs.
// it is called indirectly thru a closure function that
// supplies the methods argument.
func dispatch(methods verbMap, w http.ResponseWriter, req *http.Request) {
	log.Debugf("in dispatch: method=%s path=%s", req.Method, req.URL.Path)
	defer func() {
		_ = req.Body.Close()
	}()

	verbFunc, err := getFunc(methods, req.Method)
	if err != nil {
		errorResponse(w, err)
		return
	}

	code, data := verbFunc(req)

	rawdata, err := convData(data)
	if err != nil {
		errorResponse(w, err)
		return
	}

	w.WriteHeader(code)
	_, _ = w.Write(rawdata)

	log.Debugf("in dispatch: code=%d", code)
}

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

// errorResponse() writes to the ResponseWriter,
// the given error's message, and logs it.
func errorResponse(w http.ResponseWriter, err error) {
	code := http.StatusInternalServerError
	msg := err.Error()
	data, _ := convData(ErrorResponse{code,msg})

        w.WriteHeader(code)
        _, _ = w.Write(data)

        log.Errorf("error handling API request: %s", msg)
}

// confGet() returns the config value of the named string,
// or if there is no configured value, the given default value.
func confGet(cfg apid.ConfigService, vname string, defval string) string {
	ret := cfg.GetString(vname)
	if ret == "" {
		return defval
	}
	return ret
}

// initConfig() sets up some global configuration parameters for this plugin.
func initConfig() {
	cfg := apid.Config()

	dbName := confGet(cfg, "apidCRUD_db_name", "apidCRUD.db")
	log.Debugf("apidCRUD_db_name = %s", dbName)

	base_path := confGet(cfg, "apidCRUD_base_path", "/apid")
	log.Debugf("apidCRUD_base_path = %s", base_path)
}
