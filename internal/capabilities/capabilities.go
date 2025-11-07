package capabilities

type Capabilities uint64

const (
	cAP_NONE                   Capabilities = 0
	CAP_LOGGING                Capabilities = 1 << 1
	CAP_LOGGING_FATAL          Capabilities = 1 << 2
	CAP_HTTP_REGISTER          Capabilities = 1 << 3
	CAP_HTTP_REGISTER_WILDCARD Capabilities = 1 << 4
	CAP_HTTP_UNREGISTER        Capabilities = 1 << 5
)

func HasCapabilities(capset Capabilities, capabilities Capabilities) bool {
	return (capset & capabilities) != 0
}
