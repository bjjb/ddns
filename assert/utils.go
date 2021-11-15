package assert

import "fmt"

func repr(a interface{}) string {
	return fmt.Sprintf("%v", a)
}
