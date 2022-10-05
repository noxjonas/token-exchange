package cognito

import "errors"

var OutputType = outputAccessToken

type outputEnum string

const (
	outputJson         outputEnum = "json"
	outputAccessToken  outputEnum = "access"
	outputRefreshToken outputEnum = "refresh"
	outputIdToken      outputEnum = "id"
)

// must implement the pflag.Value interface to use the enum in cobra
//type Value interface {
//	String() string
//	Set(string) error
//	Type() string
//}

func (e *outputEnum) String() string {
	return string(*e)
}

func (e *outputEnum) Set(v string) error {
	switch v {
	case "json", "access", "refresh", "id":
		*e = outputEnum(v)
		return nil
	default:
		return errors.New(`must be one of "json", "access", "refresh", "id"`)
	}
}

func (e *outputEnum) Type() string {
	return "outputEnum"
}
