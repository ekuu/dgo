package internal

import (
	"reflect"
	"strings"

	"github.com/google/uuid"
)

func MapError[S, D any](ss []S, predicate func(int, S) (D, error)) (ds []D, err error) {
	for i, s := range ss {
		d, err := predicate(i, s)
		if err != nil {
			return nil, err
		}
		ds = append(ds, d)
	}
	return
}

func InterfaceValNil(a any) bool {
	v := reflect.ValueOf(a)
	return (v.Kind() == reflect.Ptr && v.IsNil()) || a == nil
}

func FirstUpper(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func UUIDNoHyphen() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}
