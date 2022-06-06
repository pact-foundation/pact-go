package native

import "fmt"

// Plugin Errors
var (
	ErrPluginGenericPanic             = fmt.Errorf("A general panic was caught.")
	ErrPluginMockServerStarted        = fmt.Errorf("The mock server has already been started.")
	ErrPluginInteractionHandleInvalid = fmt.Errorf("The interaction handle is invalid.	")
	ErrPluginInvalidContentType       = fmt.Errorf("The content type is not valid.")
	ErrPluginInvalidJson              = fmt.Errorf("The contents JSON is not valid JSON.")
	ErrPluginSpecificError            = fmt.Errorf("The plugin returned an error.")
)
