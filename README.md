# Welcome to gotag-validator 👋

> Generic module using this https://github.com/go-playground/validator module. This module return [problem details RFC](https://datatracker.ietf.org/doc/html/rfc7807).

## Install

```bash
	go get github.com/libercapital/gotag-validator/v2
```

## How to use

```golang
package main

import (
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	gotag_validator "github.com/libercapital/gotag-validator/v2"
	"schneider.vip/problem"
)

type MyStruct struct {
	DecimalValue string `validate:customDecimalValue`
}

var decimalValuesRegex = regexp.MustCompile("^(\\d*\\.)?\\d{2}$")

func NewValidator() (gotag_validator.IValidator, error) {
	customTags := map[string]func(fl validator.FieldLevel) bool{
		"customDecimalValue": validDecimalValue,
	}

	customMessages := map[string]string{
		"customDecimalValue": "must be in a decimal value format as %.2f",
	}

	return gotag_validator.NewValidator(customTags, customMessages)
}

func validDecimalValue(fl validator.FieldLevel) bool {
	return decimalValuesRegex.MatchString(fl.Field().String())
}

// manual call
func main() {
	myStruct := MyStruct{DecimalValue: "1234"}
	newValidator, err := gotag_validator.NewValidator(nil, nil)
	if err != nil {
		panic(err)
	}

	problemJSON := newValidator.Validate(myStruct)
	println(problemJSON)
}

// echo call
func main() {
	e := echo.New()
	myStruct := MyStruct{DecimalValue: "1234"}

	validator, err := gotag_validator.NewValidator(nil, nil)
	if err != nil {
		panic(err)
	}

	e.Validator = validator

	if err := e.Validator.Validate(myStruct); err.(*problem.Problem) != nil {
		println(err.(*problem.Problem))
	}
}
```

### Tag Global Validators

We have a global validators to use in all liber projects, for now we have only this custom tags that can be used by liber
projects.

| TAG            |                                                         Description |
| -------------- | ------------------------------------------------------------------: |
| strdatetimegte | Custom validation of date string is greater than other ex: D1 >= D2 |
| document       |                          Custom validation of CPF and CNPJ document |
| decimal2places |          Custom validation of decimal values with %.2f float format |
| brzipcode      |         Custom validation of zip code brazilian formatted 00000-000 |

## Author

👤 **jonas.gomes**

## Contributors

👤 **Eduardo Mello**

- Github: [@EduardoRMello](https://github.com/EduardoRMello)

## Show your support

Give a ⭐️ if this project helped you!

### TODO

- Added i18n

---

_This README was generated with ❤️ by [readme-md-generator](https://github.com/kefranabg/readme-md-generator)_
