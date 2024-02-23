package gotag_validator

import (
	"fmt"
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/google/go-cmp/cmp"
	"schneider.vip/problem"
)

var isRequiredFieldProblemJSON = problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
	{"name": "required_field", "reason": "required_field is a required field"},
}))

var validEnums = []string{"1", "2"}

const (
	mustBeValidEnum = "custom_field must be one of this values %s"
)

var isInvalidEnumCustomValidation = problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
	{"name": "custom_field", "reason": fmt.Sprintf(mustBeValidEnum, validEnums)},
}))

var invalidCPF = problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
	{"name": "cpf", "reason": "cpf" + documentMessage},
}))

var invalidDecimalValue = problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
	{"name": "decimal", "reason": "decimal" + decimal2placesMessage},
}))

type args struct {
	body interface{}
}

type test struct {
	name string
	args args
	want *problem.Problem
}

type Model struct {
	RequiredField string `json:"required_field" validate:"required"`
}

func TestDoValidate(t *testing.T) {
	tests := []test{
		{
			name: "IsRequired",
			args: args{body: []struct {
				RequiredField string `json:"required_field" validate:"required"`
			}{{RequiredField: ""}}},
			want: isRequiredFieldProblemJSON,
		},
		{
			name: "CannotBeEmptyArray",
			args: args{body: []struct{}{}},
			want: cannotBeEmptyArrayProblemJSON,
		},
		{
			name: "InvalidCPF",
			args: args{body: []struct {
				CPF string `json:"cpf" validate:"document"`
			}{{CPF: "12345432"}}},
			want: invalidCPF,
		},
		{
			name: "InvalidDecimalValue",
			args: args{body: []struct {
				Decimal string `json:"decimal" validate:"decimal2places"`
			}{{Decimal: "123.000"}}},
			want: invalidDecimalValue,
		},
		{
			name: "valid_and_invalid_together",
			args: args{body: []struct {
				RequiredField string `json:"required_field" validate:"required"`
				Decimal       string `json:"decimal" validate:"decimal2places"`
			}{{RequiredField: "test", Decimal: "123.000"}}},
			want: invalidDecimalValue,
		},
		{
			name: "success_struct_defined_before",
			args: args{body: Model{RequiredField: ""}},
			want: isRequiredFieldProblemJSON,
		},
		{
			name: "Success",
			args: args{body: []struct {
				Field string `json:"field" validate:"required"`
			}{{Field: "xpto"}, {Field: "xpto"}}},
		},
		{
			name: "error testing datetime gte",
			args: args{body: []struct {
				Field1 string `json:"field1"`
				Field2 string `json:"field2" validate:"strdatetimegte=field1"`
			}{{Field1: "2022-05-01", Field2: "2022-01-01"}}},
			want: problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
				{"name": "field2", "reason": "field2 must be greater than or equal field1"},
			})),
		},
		{
			name: "error testing datetime gte without json tag",
			args: args{body: []struct {
				Field1 string
				Field2 string `json:"field2" validate:"strdatetimegte=Field1"`
			}{{Field1: "2022-05-01", Field2: "2022-01-01"}}},
			want: problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
				{"name": "field2", "reason": "field2 must be greater than or equal Field1"},
			})),
		},
		{
			name: "error datetime",
			args: args{body: []struct {
				Field1 string
				Field2 string `json:"field2" validate:"strdatetimegte=Field1"`
			}{{Field1: "a", Field2: "a"}}},
			want: problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
				{"name": "field2", "reason": "field2 must be greater than or equal Field1"},
			})),
		},
		{
			name: "success string datetime greater than or equal",
			args: args{body: []struct {
				Field1 string
				Field2 string `json:"field2" validate:"strdatetimegte=Field1"`
			}{{Field1: "2022-01-01", Field2: "2022-01-01"}}},
		},
		{
			name: "invalid zip code",
			args: args{body: []struct {
				Field string `json:"field" validate:"brzipcode"`
			}{{Field: "a"}}},
			want: problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
				{"name": "field", "reason": "field must be a valid br zip code"},
			})),
		}, {
			name: "invalid iso date day",
			args: args{body: []struct {
				Field string `json:"field" validate:"iso8601date"`
			}{{Field: "2022-12-99"}}},
			want: problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
				{"name": "field", "reason": "field must be a YYYY-MM-DD format date"},
			})),
		},
		{
			name: "invalid iso date month",
			args: args{body: []struct {
				Field string `json:"field" validate:"iso8601date"`
			}{{Field: "2022-44-01"}}},
			want: problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
				{"name": "field", "reason": "field must be a YYYY-MM-DD format date"},
			})),
		},
		{
			name: "invalid iso date year",
			args: args{body: []struct {
				Field string `json:"field" validate:"iso8601date"`
			}{{Field: "000-07-01"}}},
			want: problem.New(problem.Status(400)).Append(problem.Custom("invalid_params", []map[string]interface{}{
				{"name": "field", "reason": "field must be a YYYY-MM-DD format date"},
			})),
		},
		{
			name: "valid iso date",
			args: args{body: []struct {
				Field string `json:"field" validate:"iso8601date"`
			}{{Field: "2022-12-12"}}},
			want: nil,
		},
		{
			name: "success with hifen",
			args: args{body: []struct {
				Field string `json:"field" validate:"brzipcode"`
			}{{Field: "00000-000"}}},
		},
		{
			name: "success without hifen",
			args: args{body: []struct {
				Field string `json:"field" validate:"brzipcode"`
			}{{Field: "00000000"}}},
		},
	}

	gotagValidator, _ := NewValidator(nil, nil)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gotagValidator.Validate(tt.args.body)

			if got != nil && !cmp.Equal(got.(*problem.Problem).JSONString(), tt.want.JSONString()) {
				t.Errorf("Validate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDoValidate_CustomTag(t *testing.T) {
	tests := []test{
		{
			name: "customValidation",
			args: args{body: struct {
				CustomField string `json:"custom_field" validate:"customValidationEnum"`
			}{CustomField: "invalidEnum"}},
			want: isInvalidEnumCustomValidation,
		},
	}

	customTags := map[string]func(fl validator.FieldLevel) bool{
		"customValidationEnum": validCustomEnum,
	}

	customMessages := map[string]string{
		"customValidationEnum": fmt.Sprintf(mustBeValidEnum, validEnums),
	}

	gotagValidator, _ := NewValidator(customTags, customMessages)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gotagValidator.Validate(tt.args.body)

			fmt.Println(got)

			if got != nil && !cmp.Equal(tt.want.JSONString(), got.(*problem.Problem).JSONString()) {
				t.Errorf("DoValidate() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValidator_InvalidCustomTag(t *testing.T) {
	customTags := map[string]func(fl validator.FieldLevel) bool{
		"invalidCustomTag": nil,
	}

	_, err := NewValidator(customTags, nil)
	if !cmp.Equal(err.Error(), ErrInvalidCustomValidationTag.Error()) {
		t.Errorf("NewValidator got = %v, want = %v", err, ErrInvalidCustomValidationTag)
	}
}

func validCustomEnum(fl validator.FieldLevel) bool {
	for _, element := range validEnums {
		if fl.Field().String() == element {
			return true
		}
	}

	return false
}
