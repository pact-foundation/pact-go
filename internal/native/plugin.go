package native

import "errors"

// Plugin Errors.
var (
	ErrPluginGenericPanic             = errors.New("a general panic was caught")
	ErrPluginMockServerStarted        = errors.New("the mock server has already been started")
	ErrPluginInteractionHandleInvalid = errors.New("the interaction handle is invalid")
	ErrPluginInvalidContentType       = errors.New("the content type is not valid")
	ErrPluginInvalidJson              = errors.New("the contents JSON is not valid JSON")
	ErrPluginSpecificError            = errors.New("the plugin returned an error")
)
