package compare

import (
	"github.com/gobwas/glob"
	"github.com/tidwall/gjson"
	"io"
	"strings"
)

type Func[T any] func(T) (bool, error)

func Glob(pattern string) Func[string] {
	return func(s2 string) (bool, error) {
		g, err := glob.Compile(pattern)
		if err != nil {
			return false, err
		}
		return g.Match(s2), nil
	}
}

func HasPrefix(s string) Func[string] {
	return func(s2 string) (bool, error) {
		return strings.HasPrefix(s2, s), nil
	}
}

func Exactly(s string) Func[string] {
	return func(s2 string) (bool, error) {
		return s2 == s, nil
	}
}

func Contains(s string) Func[string] {
	return func(s2 string) (bool, error) {
		return strings.Contains(s2, s), nil
	}
}

func JSON[T any](key string, contains Func[T]) Func[io.Reader] {
	return func(r io.Reader) (bool, error) {
		data, err := io.ReadAll(r)
		if err != nil {
			return false, err
		}

		if !gjson.ValidBytes(data) {
			return false, nil
		}

		result := gjson.GetBytes(data, key)

		val, ok := result.Value().(T)
		if !ok {
			return false, nil
		}

		return contains(val)
	}
}
