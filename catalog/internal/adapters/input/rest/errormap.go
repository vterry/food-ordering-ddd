package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator"
	common "github.com/vterry/food-ordering/common/pkg"
)

func httpStatusFor(err error) int {

	var notFound common.NotFoundErr
	var bizRule common.BusinessRuleErr
	var valErr common.ValidationErr

	var validationErrors validator.ValidationErrors
	var syntaxError *json.SyntaxError

	switch {
	case errors.As(err, &notFound):
		return http.StatusNotFound
	case errors.As(err, &bizRule):
		return http.StatusUnprocessableEntity
	case errors.As(err, &valErr):
		return http.StatusBadRequest
	case errors.As(err, &validationErrors):
		return http.StatusBadRequest
	case errors.As(err, &syntaxError):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}

}
