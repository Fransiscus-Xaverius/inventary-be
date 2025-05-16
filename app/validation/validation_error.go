package validation

type ValidationError struct {
	Error      string `json:"error"`
	ErrorField string `json:"error_column"`
}
