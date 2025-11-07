#include "cffi.h"

#include <string.h>
#include <stdio.h>
#include <stdint.h>
#include <stdbool.h>

static or_api_t api = {
    .version  = MODLOADER_VERSION,
    .muid     = 0,
    .loginfo  = loginfo,
    .logwarn  = logwarn,
    .logerror = logerror,
    .logfatal = logfatal,
    .register_http = or_register_http,
    .unregister_http = or_unregister_http
};

static loadmod_err_t error_reg;

bool cffi_health(void) {
#ifdef __WIN32__
    return true;
#elif __linux__
    return true;
#elif __APPLE__
    return false;
#elif __UNIX__
    return false;
#else
    return false;
#endif
}

loadmod_err_t get_error(void) {
    return error_reg;
}

inline static void set_error(loadmod_err_t error) {
    error_reg = error;
}

void call_or_http_handler(or_http_handler_t fn, or_ctx_t* ctx, or_http_req_t* req, void* extra) {
    fn(ctx, req, extra);
}

inline static void cffi_create_api_struct(or_api_t* module_api, muid_t muid) {
    memcpy(module_api, &api, sizeof(or_api_t));
    module_api->muid = muid;
}

#define INIT_FUNC_FAIL "Warning: init function for \"%s\" returned false (failed state)"

inline static loadmod_err_t cffi_common_init_call(char* path, init_func_t init_func, muid_t muid) {
    loadmod_err_t ret = LOADMOD_SUCCESS;

    or_api_t module_api = { 0 };
    cffi_create_api_struct(&module_api, muid);

    /* Call init function */
    bool success = init_func(&module_api);
    if (!success) {
        uint32_t len = strlen(path) + sizeof(INIT_FUNC_FAIL);
        char* buf = alloca(len);
        snprintf(buf, len, INIT_FUNC_FAIL, path);
        log_warn(buf);
        ret = LOADMOD_INIT_FUNC_STATE_FAIL;
    }

    return ret;
}


#ifdef __linux__

#define LOAD_SO_ERROR_MSG "Invalid module path: %s"
#define DLSYM_ERROR_MSG "dlsym() error during \"init\" function loading: %s"
#define DLSYM_UNINIT_ERROR_MSG "dlsym() error during \"uninit\" function loading: %s"
#define DLCLOSE_ERROR_MSG "dlclose() error when closing module!"

inline static mod_handle_t cffi_load_so(char* path, muid_t muid) {
    void* handle = dlopen(path, RTLD_NOW);
    if (!handle) {
        uint32_t len = strlen(path) + sizeof(LOAD_SO_ERROR_MSG);
        char* buf = alloca(len);
        snprintf(buf, len, LOAD_SO_ERROR_MSG, path);
        log_error(buf);
        set_error(LOADMOD_NO_SUCH_MOD);
        return NULL;
    }

    /* Clear errors */
    dlerror();

    /* Load init() function */
    init_func_t init_func = (init_func_t) dlsym(handle, "init");
    char* error = dlerror();
    if (error != NULL) {
        uint32_t len = strlen(error) + sizeof(DLSYM_ERROR_MSG);
        char* buf = alloca(len);
        snprintf(buf, len, DLSYM_ERROR_MSG, error);
        log_error(buf);
        set_error(LOADMOD_NO_VALID_INIT_FUNC);
        return handle;
    }

    set_error(cffi_common_init_call(path, init_func, muid));
    return handle;
}

inline static void cffi_unload_so(mod_handle_t handle) {
    uninit_func_t uninit_func = (uninit_func_t) dlsym(handle, "uninit");
    char* error = dlerror();
    if (error != NULL) {
        uint32_t len = strlen(error) + sizeof(DLSYM_UNINIT_ERROR_MSG);
        char* buf = alloca(len);
        snprintf(buf, len, DLSYM_UNINIT_ERROR_MSG, error);
        log_error(buf);
        set_error(LOADMOD_NO_VALID_UNINIT_FUNC);
    } else {
        uninit_func(&api);
    }

    if (dlclose(handle) != 0) {
        log_error(DLCLOSE_ERROR_MSG);
        set_error(LOADMOD_CLOSE_FAIL);

    }
}

#elif defined(_WIN32)

#define LOAD_DLL_ERROR_MSG "Invalid module path: %s"
#define GETPROCADDRESS_ERROR_MSG "GetProcAddress() error during \"init\" function loading (ERRNR: 0x%08x): %s"
#define GETPROCADDRESS_UNINIT_ERROR_MSG "GetProcAddress() error during \"uninit\" function loading (ERRNR: 0x%08x): %s"
#define FREE_DLL_ERROR_MSG "Unable to unload module with handle: 0x%08x"
#define MAX_UINT64_HEX_LEN 8

inline static mod_handle_t cffi_load_dll(char* path, muid_t muid) {
    HMODULE handle = LoadLibraryExA(path, NULL, 0x0);
    if (handle == NULL) {
        DWORD error_nr = GetLastError();
        char *msg = NULL;
        uint32_t len = strlen(path) + sizeof(LOAD_DLL_ERROR_MSG);
        char* buf = _alloca(len);
        _snprintf(buf, len, LOAD_DLL_ERROR_MSG, path);
        log_error(buf);
        set_error(LOADMOD_NO_SUCH_MOD);
        return NULL;
    }

    init_func_t init_func = (init_func_t) GetProcAddress(handle, "init");
    if (init_func == NULL) {
        DWORD error_nr = GetLastError();
        char *msg = NULL;
        FormatMessageA(FORMAT_MESSAGE_ALLOCATE_BUFFER | FORMAT_MESSAGE_FROM_SYSTEM
             | FORMAT_MESSAGE_IGNORE_INSERTS, NULL, error_nr, 0, (LPSTR)&msg, 0, NULL);
        uint32_t len = strlen(msg) + sizeof(GETPROCADDRESS_ERROR_MSG) + MAX_UINT64_HEX_LEN;
        char* buf = _alloca(len);
        _snprintf(buf, len, GETPROCADDRESS_ERROR_MSG, (unsigned)error_nr, msg ? msg : "unknown");
        log_error(buf);
        if (msg) LocalFree(msg);
        set_error(LOADMOD_NO_VALID_INIT_FUNC);
        return handle;
    }

    set_error(cffi_common_init_call(path, init_func, muid));
    return handle;
}

inline static void cffi_unload_dll(mod_handle_t handle) {
    uninit_func_t uninit_func = (uninit_func_t) GetProcAddress(handle, "uninit");
    if (uninit_func == NULL) {
        DWORD error_nr = GetLastError();
        char *msg = NULL;
        FormatMessageA(FORMAT_MESSAGE_ALLOCATE_BUFFER | FORMAT_MESSAGE_FROM_SYSTEM
             | FORMAT_MESSAGE_IGNORE_INSERTS, NULL, error_nr, 0, (LPSTR)&msg, 0, NULL);
        uint32_t len = strlen(msg) + sizeof(GETPROCADDRESS_UNINIT_ERROR_MSG) + MAX_UINT64_HEX_LEN;
        char* buf = _alloca(len);
        _snprintf(buf, len, GETPROCADDRESS_UNINIT_ERROR_MSG, (unsigned)error_nr, msg ? msg : "unknown");
        log_error(buf);
        if (msg) LocalFree(msg);
        set_error(LOADMOD_NO_VALID_UNINIT_FUNC);
    } else {
        uninit_func(&api);
    }

    if (FreeLibrary(handle) == false) {
        DWORD error_nr = GetLastError();
        char *msg = NULL;
        uint32_t len = MAX_UINT64_HEX_LEN + sizeof(FREE_DLL_ERROR_MSG);
        char* buf = _alloca(len);
        _snprintf(buf, len, FREE_DLL_ERROR_MSG, handle);
        log_error(buf);
        set_error(LOADMOD_CLOSE_FAIL);
    }
}



#endif

mod_handle_t cffi_load_module(char* path, muid_t muid) {
    #ifdef __linux__
        return cffi_load_so(path, muid);
    #elif _WIN32
        return cffi_load_dll(path, muid);
    #else
        log_error("Unsupported OS detected!");
    #endif

    set_error(LOADMOD_UNSUPPORTED_OS);
    return NULL;
}

void cffi_unload_module(mod_handle_t handle) {
    #ifdef __linux__
        cffi_unload_so(handle);
    #elif _WIN32
        cffi_unload_dll(handle);
    #else
        log_error("Unsupported OS detected!");
    #endif

    set_error(LOADMOD_UNSUPPORTED_OS);
}
