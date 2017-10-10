package apidCRUD

import "github.com/apid/apid-core"

// init() is magically called at startup by the go runtime.
// we take this opportunity to tell apid to call our initPlugin()
// function when it does InitializePlugins().
func init() {
	apid.RegisterPlugin(initPlugin, pluginData)
}
