package gotag_validator

import (
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	"schneider.vip/problem"
)

type InvalidParam struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

const (
	customPrefix               = "custom"
	regexForValidDecimalValues = "^(\\d*\\.)?\\d{2}$"
	regexForValidBrZipCode     = "^\\d{5}\\-{0,1}\\d{3}$"

	datetimeGteTag    = "strdatetimegte"
	documentTag       = "document"
	decimal2placesTag = "decimal2places"
	brZipCodeTag      = "brzipcode"
	iso8601dateTag    = "iso8601date"

	mustBeGteMessage      = " must be greater than or equal "
	decimal2placesMessage = " must be in a decimal value format as %.2f"
	documentMessage       = " must be a valid cpf or cnpj"
	zipCodeMessage        = " must be a valid br zip code"
	isoMessage            = " must be a YYYY-MM-DD format date"
)

var decimalValuesRegex = regexp.MustCompile(regexForValidDecimalValues)
var regexBrZipCode = regexp.MustCompile(regexForValidBrZipCode)

type IValidator interface {
	Validate(body interface{}) error
}

// ErrInvalidCustomValidationTag : Error that represent invalid custom validation tag.
var ErrInvalidCustomValidationTag = errors.New("the custom tag must starts with 'custom' characters. (i.e validate:\"customMyValidation\")")
var cannotBeEmptyArrayProblemJSON = problem.New(problem.Status(http.StatusBadRequest), problem.Title("Cannot be empty."))

type genericValidator struct {
	tagValidator  *validator.Validate
	trans         ut.Translator
	messageErrors map[string]string //must be key - tagName
}

// NewValidator : Return a instance of IValidator interface, that impl a generic interface validate, based on tags `i.e validate:"customValidation"
// if the key of customValidation is not starts with 'custom*' then return a ErrInvalidCustomValidationTag
// This method register the EN language default
// TODO: Added i18n getting environment client of this module.
func NewValidator(customValidations map[string]func(fl validator.FieldLevel) bool, messageErrors map[string]string) (IValidator, error) {
	tagValidator := validator.New()

	for k, v := range customValidations {
		if !strings.HasPrefix(k, customPrefix) {
			return nil, ErrInvalidCustomValidationTag
		}

		if err := tagValidator.RegisterValidation(k, v); err != nil {
			return nil, err
		}
	}
	if err := tagValidator.RegisterValidation(datetimeGteTag, validateDatetimeGTE); err != nil {
		return nil, err
	}
	if err := tagValidator.RegisterValidation(documentTag, validDocument); err != nil {
		return nil, err
	}
	if err := tagValidator.RegisterValidation(decimal2placesTag, validDecimalValue); err != nil {
		return nil, err
	}
	if err := tagValidator.RegisterValidation(brZipCodeTag, validateBrZipCode); err != nil {
		return nil, err
	}
	if err := tagValidator.RegisterValidation(iso8601dateTag, validateISO8601Date); err != nil {
		return nil, err
	}

	trans, err := registerDefaultTranslation(tagValidator)
	if err != nil {
		return nil, err
	}

	tagValidator.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	return &genericValidator{tagValidator, trans, messageErrors}, nil
}

func registerDefaultTranslation(tagValidator *validator.Validate) (ut.Translator, error) {
	localeEN := en.New()
	uni := ut.New(localeEN, localeEN)
	trans, _ := uni.GetTranslator("en")

	return trans, en_translations.RegisterDefaultTranslations(tagValidator, trans)
}

func validateDatetimeGTE(fl validator.FieldLevel) bool {
	fieldValue := fl.Field().String()

	otherFieldName := fl.Param()
	var otherFieldVal reflect.Value
	if fl.Parent().Kind() == reflect.Ptr {
		otherFieldVal = fl.Parent().Elem().FieldByName(otherFieldName)
	} else {
		otherFieldVal = fl.Parent().FieldByName(otherFieldName)
	}

	referenceField, err := parseToUTCTime(fieldValue)
	if err != nil {
		return false
	}

	compareField, err := parseToUTCTime(otherFieldVal.String())
	if err != nil {
		return false
	}

	if referenceField.Before(compareField.Time) {
		return false
	}

	return true
}

func validDocument(fl validator.FieldLevel) bool {
	document := fl.Field().String()

	if isCPF(document) {
		return true
	}

	return isCNPJ(document)
}

func validDecimalValue(fl validator.FieldLevel) bool {
	return decimalValuesRegex.MatchString(fl.Field().String())
}

func validateBrZipCode(fl validator.FieldLevel) bool {
	return regexBrZipCode.MatchString(fl.Field().String())
}

func validateISO8601Date(fl validator.FieldLevel) bool {
	_, err := time.Parse("2006-01-02", fl.Field().String())
	return err == nil
}

// DoValidate : Reflection validation, can be a []Slice of []Array or simple Struct{} or Interface{}
// Will Return *problem.Problem if has any problem, or will return nil if all is alright
func (gv genericValidator) Validate(body interface{}) error {
	kind := reflect.TypeOf(body).Kind()
	switch kind {
	case reflect.Slice:
		slice := reflect.ValueOf(body)
		if slice.Len() == 0 {
			return cannotBeEmptyArrayProblemJSON
		}

		for i := 0; i < slice.Len(); i++ {
			if problemJSON := gv.validate(slice.Index(i).Interface()); problemJSON != nil {
				return problemJSON
			}

			continue
		}
	case reflect.Struct, reflect.Interface, reflect.Ptr:
		return gv.validate(body)
	}

	return nil
}

func (gv genericValidator) validate(body interface{}) *problem.Problem {
	if err := gv.tagValidator.Struct(body); err != nil {
		validationErrors := err.(validator.ValidationErrors)

		if len(validationErrors) > 0 {
			invalidParams := make([]InvalidParam, len(validationErrors))

			for index, validationError := range validationErrors {
				paths := strings.Split(validationError.Namespace(), ".")

				//fix to not show principal struct name, only children
				if len(paths) > 1 {
					invalidParams[index].Name = strings.Join(paths[1:], ".")
				} else {
					invalidParams[index].Name = validationError.Namespace()
				}

				switch {
				case strings.HasPrefix(validationError.Tag(), customPrefix):
					for customTag, message := range gv.messageErrors {
						if validationError.Tag() == customTag {
							invalidParams[index].Reason = message
							break
						}
					}
				case validationError.Tag() == "required_without_all":
					tags := getJsonTagNameFromPropertyName(validationError.Param(), body)

					invalidParams[index].Reason = fmt.Sprintf("at least one of these fields are required: %v, %v", validationError.Field(), strings.Join(tags, ", "))
				case validationError.Tag() == "required_with":
					tags := getJsonTagNameFromPropertyName(validationError.Param(), body)

					invalidParams[index].Reason = fmt.Sprintf("%v is required if one of these fields are filled: %v", validationError.Field(), strings.Join(tags, ", "))
				case validationError.Tag() == datetimeGteTag:
					tag := getJsonTagNameFromPropertyName(validationError.Param(), body)[0]

					invalidParams[index].Reason = validationError.Field() + mustBeGteMessage + tag
				case validationError.Tag() == documentTag:
					invalidParams[index].Reason = validationError.Field() + documentMessage
				case validationError.Tag() == decimal2placesTag:
					invalidParams[index].Reason = validationError.Field() + decimal2placesMessage
				case validationError.Tag() == brZipCodeTag:
					invalidParams[index].Reason = validationError.Field() + zipCodeMessage
				case validationError.Tag() == iso8601dateTag:
					invalidParams[index].Reason = validationError.Field() + isoMessage
				default:
					invalidParams[index].Reason = validationError.Translate(gv.trans)
				}
			}

			return problem.New(problem.Status(http.StatusBadRequest)).Append(problem.Custom("invalid_params", invalidParams))
		}
	}

	return nil
}

func getJsonTagNameFromPropertyName(param string, body interface{}) (jsonNames []string) {
	pointr := reflect.TypeOf(body)

	for _, field := range strings.Split(param, " ") {

		var otherFieldVal reflect.StructField
		if pointr.Kind() == reflect.Ptr {
			otherFieldVal, _ = pointr.Elem().FieldByName(field)
		} else {
			otherFieldVal, _ = pointr.FieldByName(field)
		}

		tag := ""
		if tag = otherFieldVal.Tag.Get("json"); tag == "" {
			if tag = otherFieldVal.Tag.Get("param"); tag == "" {
				if tag = otherFieldVal.Tag.Get("query"); tag == "" {
					tag = field
				}
			}
		}

		jsonNames = append(jsonNames, strings.Split(tag, ",")[0])
	}

	return
}

type UTCTime struct {
	time.Time
}

const defaultFormat = time.RFC3339
const MilisecondsFormat = "2006-01-02T15:04:05.000Z"
const DateTimeZeroFormat = "2006-01-02T15:04:05Z"

var layouts = []string{
	defaultFormat,
	"2006-01-02T15:04Z",       // ISO 8601 UTC
	DateTimeZeroFormat,        // ISO 8601 UTC
	MilisecondsFormat,         // ISO 8601 UTC
	"2006-01-02T15:04:05",     // ISO 8601 UTC
	"2006-01-02 15:04",        // Custom UTC
	"2006-01-02 15:04:05",     // Custom UTC
	"2006-01-02 15:04:05.000", // Custom UTC
	"2006-01-02",
}

func parseToUTCTime(timeString string) (utc UTCTime, err error) {
	var parsed time.Time
	for _, layout := range layouts {
		parsed, err = time.Parse(layout, timeString)
		if err == nil {
			utc.Time = parsed.UTC()
			return
		}
	}
	return utc, fmt.Errorf("invalid date format: %s", timeString)
}
