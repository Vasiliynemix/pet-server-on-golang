package handlers

import "github.com/go-playground/validator/v10"

type ValidationError struct {
	Field string `json:"field"`
	Tag   string `json:"tag"`
	Value string `json:"value"`
}

func CreateValidationErrorsResp(req interface{}) []*ValidationError {
	var errs []*ValidationError

	err := validator.New().Struct(req)
	if err != nil {
		for _, validErr := range err.(validator.ValidationErrors) {
			var element ValidationError
			element.Field = validErr.StructNamespace()
			element.Tag = validErr.Tag()
			element.Value = validErr.Param()
			errs = append(errs, &element)
		}
	}
	return errs
}
