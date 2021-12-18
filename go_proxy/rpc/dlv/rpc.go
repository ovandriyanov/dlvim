package dlv

import (
	"fmt"
)

const ServiceName = "RPCServer" // Dlv client expects that service name

func FQMN(method string) string {
	return fmt.Sprintf("%s.%s", ServiceName, method)
}
