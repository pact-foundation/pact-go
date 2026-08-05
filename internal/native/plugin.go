package native

import "fmt"

// Plugin Errors
var (
	ErrPluginGenericPanic             = fmt.Errorf("a general panic was caught")
	ErrPluginMockServerStarted        = fmt.Errorf("the mock server has already been started")
	ErrPluginInteractionHandleInvalid = fmt.Errorf("the interaction handle is invalid")
	ErrPluginInvalidContentType       = fmt.Errorf("the content type is not valid")
	ErrPluginInvalidJson              = fmt.Errorf("the contents JSON is not valid JSON")
	ErrPluginSpecificError            = fmt.Errorf("the plugin returned an error")
)
