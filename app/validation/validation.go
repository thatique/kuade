package validation

type Result struct {
	Ok     bool
	Err    error
	Errors []FieldError
}

type FieldError struct {
	Field   string `xml:"Field" json:"field"`
	Message string `xml:"Message" json:"message"`
}

func (r *Result) AddFieldError(field string, messages ...string) {
	if r.Errors == nil {
		r.Errors = make([]FieldError, 0)
	}

	for _, message := range messages {
		r.Errors = append(r.Errors, FieldError{
			Field:   field,
			Message: message,
		})
		r.Ok = false
	}
}

func Success() *Result {
	return &Result{Ok: true}
}

// Failed returns a failed validation result
func Failed(messages ...string) *Result {
	r := &Result{Ok: false}
	r.AddFieldError("", messages...)
	return r
}

// Error returns a failed validation result
func Error(err error) *Result {
	return &Result{Ok: false, Err: err}
}
