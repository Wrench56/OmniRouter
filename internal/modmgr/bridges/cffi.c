#include "cffi.h"

#include <string.h>
#include <stdio.h>
#include <stdint.h>
#include <stdbool.h>

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


#define INIT_VERSION_LOWER_WRN "Version of module \"%s\" (v%d) is lower than current module loader version (v%d). Continuing without guarantees..."
#define INIT_STRUCT_ERROR_MSG "init_struct returned by module %s was NULL"

inline static loadmod_err_t cffi_common_init_call(char* path, init_func_t init_func) {
    /* Call init function and save init_struct */
    init_struct_t* init_struct = init_func();
    if (init_struct == NULL) {
        uint32_t len = strlen(path) + sizeof(INIT_STRUCT_ERROR_MSG);
        char* buf = alloca(len);
        snprintf(buf, len, INIT_STRUCT_ERROR_MSG, path);
        log_error(buf);
        return LOADMOD_INIT_STRUCT_NULL;
    }

    /* init_struct version mismatch check */
    if (init_struct->version < MODLOADER_VERSION) {
        uint32_t len = strlen(path) + sizeof(INIT_VERSION_LOWER_WRN) + MAX_VERSION_LENGTH * 2;
        char* buf = alloca(len);
        snprintf(buf, len, INIT_VERSION_LOWER_WRN, path, init_struct->version, MODLOADER_VERSION);
        log_warn(buf);
    }

    return LOADMOD_SUCCESS;
}


#ifdef __linux__

#define LOAD_SO_ERROR_MSG "Invalid module path: %s"
#define DLSYM_ERROR_MSG "dlsym() error during \"init\" function loading: %s"

inline static loadmod_err_t cffi_load_so(char* path) {
    void* handle = dlopen(path, RTLD_NOW);
    if (!handle) {
        uint32_t len = strlen(path) + sizeof(LOAD_SO_ERROR_MSG);
        char* buf = alloca(len);
        snprintf(buf, len, LOAD_SO_ERROR_MSG, path);
        log_error(buf);
        return LOADMOD_NO_SUCH_MOD;
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
        return LOADMOD_NO_VALID_INIT_FUNC;
    }

    return cffi_common_init_call(path, init_func);
}
#elif defined(__WIN32__)

#define LOAD_DLL_ERROR_MSG "Invalid module path: %s"
#define GETPROCADDRESS_ERROR_MSG "GetProcAddress() error during \"init\" function loading (ERRNR: 0x%08x): %s"
#define MAX_UINT64_HEX_LEN 8

inline static loadmod_err_t cffi_load_dll(char* path) {
    HMODULE handle = LoadLibraryExA(path, NULL, 0x0);
    if (handle == NULL) {
        DWORD error_nr = GetLastError();
        char *msg = NULL;
        uint32_t len = strlen(path) + sizeof(LOAD_DLL_ERROR_MSG);
        char* buf = _alloca(len);
        _snprintf(buf, len, LOAD_DLL_ERROR_MSG, path);
        log_error(buf);
        return LOADMOD_NO_SUCH_MOD;
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
        return LOADMOD_NO_VALID_INIT_FUNC;
    }

    return cffi_common_init_call(path, init_func);
}

#endif

inline loadmod_err_t cffi_load_module(char* path) {
    #ifdef __linux__
        return cffi_load_so(path);
    #elif __WIN32__
        return cffi_load_dll(path);
    #else
        log_error("Unsupported OS detected!");
    #endif

    return LOADMOD_UNSUPPORTED_OS;
}
