package container

import "fmt"

type Environment map[string]string

func (e Environment) ToSlice() (result []string) {
	for k, v := range e {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return
}
