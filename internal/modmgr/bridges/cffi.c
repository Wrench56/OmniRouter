#include "cffi.h"

#include <string.h>
#include <stdio.h>
#include <stdint.h>
#include <stdbool.h>

bool cffi_health(void) {
#ifdef __WINDOWS__
    return false;
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

#ifdef __linux__

#define LOAD_SO_ERROR_MSG "Invalid module path: %s"
#define DLSYM_ERROR_MSG "dlsym() error during \"init\" function loading: %s"
#define INIT_VERSION_LOWER_WRN "Version of module \"%s\" (v%d) is lower than current module loader version (v%d). Continuing without guarantees..."
#define INIT_STRUCT_ERROR_MSG "init_struct returned by module %s was NULL"

loadmod_err_t cffi_load_so(char* path) {
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
#endif // __linux__


loadmod_err_t cffi_load_module(char* path) {
    #ifdef __linux__
        return cffi_load_so(path);
    #else
        log_error("Unsupported OS detected!");
    #endif

    return LOADMOD_UNSUPPORTED_OS;
}
