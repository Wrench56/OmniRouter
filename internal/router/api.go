package router

import (
	"omnirouter/internal/capabilities"
	"omnirouter/internal/logger"
	"strings"
)

const (
	SUCCESS = 0
	/* Do not use error code 1! */
	ERR_FFI_RESERVED = 1
	ERR_REG_CAP      = 2
	ERR_REG_WILD_CAP = 3
	ERR_UNREG_CAP    = 2
)

const methodCount = 7
const (
	METHOD_UNKNOWN uint8 = 0
	METHOD_GET     uint8 = 1 << 1
	METHOD_HEAD    uint8 = 1 << 2
	METHOD_POST    uint8 = 1 << 3
	METHOD_PUT     uint8 = 1 << 4
	METHOD_DELETE  uint8 = 1 << 5
	METHOD_PATCH   uint8 = 1 << 6
	METHOD_OPTIONS uint8 = 1 << 7
	METHOD_ANY     uint8 = ^uint8(0)
)

func (r *radixRouter) Register(caps capabilities.Capabilities, methodMask uint8, path string, h HTTPHandler) uint64 {
	if !capabilities.HasCapabilities(caps, capabilities.CAP_HTTP_REGISTER) {
		logger.Warn("Insufficient capabilities to register an HTTP route",
			"capabilities", caps, "needed", capabilities.CAP_HTTP_REGISTER)
		return ERR_REG_CAP
	}

	p := path
	isWildcard := false
	if strings.HasSuffix(p, "*") {
		if !capabilities.HasCapabilities(caps, capabilities.CAP_HTTP_REGISTER_WILDCARD) {
			logger.Warn("Insufficient capabilities to register a wildcard HTTP route",
				"capabilities", caps,
				"needed", capabilities.CAP_HTTP_REGISTER_WILDCARD&capabilities.CAP_HTTP_REGISTER)
			return ERR_REG_WILD_CAP
		}

		isWildcard = true
	}

	p = normalize(p)
	re := r.getOrCreate(p)
	r.mu.Lock()
	defer r.mu.Unlock()

	if isWildcard {
		re.wildcard = true
	}

	execForMethodBit(func(i int) {
		re.table.Handlers[i] = h
	}, uint8(methodMask))

	logger.Info("Added/updated HTTP handler", "path", p, "method_mask", methodMask)
	return SUCCESS
}

func (r *radixRouter) Unregister(caps capabilities.Capabilities, methodMask uint8, path string) uint64 {
	if !capabilities.HasCapabilities(caps, capabilities.CAP_HTTP_UNREGISTER) {
		logger.Warn("Insufficient capabilities to unregister an HTTP route",
			"capabilities", caps, "needed", capabilities.CAP_HTTP_UNREGISTER)
		return ERR_UNREG_CAP
	}

	p := normalize(path)

	r.mu.Lock()
	defer r.mu.Unlock()

	v, ok := r.tree.Get(p)
	if !ok {
		logger.Info("Unregister called on missing path", "path", p)
		return SUCCESS
	}

	re := v.(*routeEntry)
	execForMethodBit(func(i int) {
		re.table.Handlers[i] = nil
	}, methodMask)

	logger.Info("Unregistered HTTP handler", "path", p, "method_mask", methodMask)
	return SUCCESS
}
