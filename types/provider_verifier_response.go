package types

// ProviderVerifierResponse contains the output of the pact-provider-verifier
// command.
type ProviderVerifierResponse struct {
	Version  string `json:"version"`
	Examples []struct {
		ID              string      `json:"id"`
		Description     string      `json:"description"`
		FullDescription string      `json:"full_description"`
		Status          string      `json:"status"`
		FilePath        string      `json:"file_path"`
		LineNumber      int         `json:"line_number"`
		RunTime         float64     `json:"run_time"`
		PendingMessage  interface{} `json:"pending_message"`
		Mismatches      []string    `json:"mismatches"`
		Pact            struct {
			ConsumerName     string `json:"consumer_name"`
			ProviderName     string `json:"provider_name"`
			URL              string `json:"url"`
			ShortDescription string `json:"short_description"`
		} `json:"pact"`
		Exception struct {
			Class     string   `json:"class"`
			Message   string   `json:"message"`
			Backtrace []string `json:"backtrace"`
		} `json:"exception,omitempty"`
	} `json:"examples"`
	Summary struct {
		Duration                     float64 `json:"duration"`
		ExampleCount                 int     `json:"example_count"`
		FailureCount                 int     `json:"failure_count"`
		PendingCount                 int     `json:"pending_count"`
		ErrorsOutsideOfExamplesCount int     `json:"errors_outside_of_examples_count"`
		Notices                      []struct {
			Text string `json:"text"`
			When string `json:"when"`
		} `json:"notices"`
	} `json:"summary"`
	SummaryLine string `json:"summary_line"`
}
