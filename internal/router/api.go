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

func (r *radixRouter) Register(caps capabilities.Capabilities, path string, h HTTPHandler) uint64 {
	if !capabilities.HasCapabilities(caps, capabilities.CAP_HTTP_REGISTER) {
		logger.Warn("Insufficient capabilities to register an HTTP route", "capabilities", caps, "needed", capabilities.CAP_HTTP_REGISTER)
		return ERR_REG_CAP
	}

	p := normalize(path)

	if strings.HasSuffix(p, "*") {
		if !capabilities.HasCapabilities(caps, capabilities.CAP_HTTP_REGISTER_WILDCARD) {

			logger.Warn("Insufficient capabilities to register a wildcard HTTP route", "capabilities", caps, "needed", capabilities.CAP_HTTP_REGISTER_WILDCARD&capabilities.CAP_HTTP_REGISTER)
			return ERR_REG_WILD_CAP
		}
		p = strings.TrimSuffix(p, "*")
		if !strings.HasSuffix(p, "/") {
			p += "/"
		}
	}

	r.mu.Lock()
	r.tree.Insert(p, h)
	r.mu.Unlock()
	logger.Info("Added new HTTP handler for path ", "path", path)
	return SUCCESS
}

func (r *radixRouter) Unregister(caps capabilities.Capabilities, path string) uint64 {
	if !capabilities.HasCapabilities(caps, capabilities.CAP_HTTP_UNREGISTER) {
		logger.Warn("Insufficient capabilities to unregister an HTTP route", "capabilities", caps, "needed", capabilities.CAP_HTTP_REGISTER)
		return ERR_UNREG_CAP
	}
	p := normalize(path)

	r.mu.Lock()
	r.tree.Delete(p)
	r.mu.Unlock()
	logger.Info("Unregistered HTTP handler for path", "path", path)
	return SUCCESS
}
