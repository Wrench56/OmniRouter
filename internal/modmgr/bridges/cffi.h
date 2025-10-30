#ifndef CFFI_H
#define CFFI_H

#include <stdint.h>
#include <stdbool.h>

#define MODLOADER_VERSION 3
#define MAX_VERSION_LENGTH 20

/* Exported functions from logger_cffi.go */
/* Do NOT use these directly. Use logger helper instead */
extern void loginfo(char* msg, char* module_);
extern void logwarn(char* msg, char* module_);
extern void logerror(char* msg, char* module_);
extern void logfatal(char* msg, char* module_);

/* `module:line` concat util */
#define S1(x) #x
#define S2(x) S1(x)
#define LOCATION __FILE__ ":" S2(__LINE__)

/* Logger helpers */
#define log_info(msg) loginfo(msg, LOCATION)
#define log_warn(msg) logwarn(msg, LOCATION)
#define log_error(msg) logerror(msg, LOCATION)
#define log_fatal(msg) logfatal(msg, LOCATION)

/* Load module error enum */
typedef enum {
    LOADMOD_SUCCESS = 0,
    LOADMOD_UNSUPPORTED_OS,
    LOADMOD_NO_SUCH_MOD,
    LOADMOD_CLOSE_FAIL,
    LOADMOD_NO_VALID_INIT_FUNC,
    LOADMOD_NO_VALID_UNINIT_FUNC,
    LOADMOD_INIT_FUNC_STATE_FAIL
} loadmod_err_t;

typedef struct {

} or_ctx_t;

typedef struct {

} or_http_req_t;

typedef void (*or_http_handler_t)(
    or_ctx_t* ctx,
    or_http_req_t* req,
    void* extra
);

typedef struct {
    uint64_t version;
    void (*loginfo)(char* msg, char* module_);
    void (*logwarn)(char* msg, char* module_);
    void (*logerror)(char* msg, char* module_);
    void (*logfatal)(char* msg, char* module_);
    uint64_t (*register_http)(char* path, or_http_handler_t handler, void* extra);
    void (*unregister_http)(char* path);
} or_api_t;

typedef struct {
    uint32_t version;
} health_struct_t;

typedef bool (*init_func_t)(const or_api_t* api);
typedef bool (*uninit_func_t)(const or_api_t* api);

/* Exported functions from modmgr.go */
extern uint64_t or_register_http(char* path, or_http_handler_t handler, void* extra);
extern void or_unregister_http(char* path);

#ifdef __linux__
    #include <dlfcn.h>
    #include <alloca.h>
    
    typedef void* mod_handle_t;

#elif _WIN32
    #include <malloc.h>
    #include <libloaderapi.h>
    #include <errhandlingapi.h>
    #include <windows.h>
    #include <winbase.h>

    typedef HMODULE mod_handle_t;
#endif

/* cffi.c exports */
bool cffi_health(void);
mod_handle_t cffi_load_module(char* path);
mod_handle_t cffi_unload_module(mod_handle_t handle);
void call_or_http_handler(or_http_handler_t fn, or_ctx_t* ctx, or_http_req_t* req, void* extra);
loadmod_err_t get_error(void);

#endif // CFFI_H
