#ifndef CFFI_H
#define CFFI_H

#include <stdint.h>
#include <stdbool.h>

#define MODLOADER_VERSION 4
#define MAX_VERSION_LENGTH 20

/* Exported functions from logger_cffi.go */
/* Do NOT use these directly. Use logger helper instead */
extern void or_loginfo(char* msg, char* module_);
extern void or_logwarn(char* msg, char* module_);
extern void or_logerror(char* msg, char* module_);
extern void or_logfatal(char* msg, char* module_);

/* `module:line` concat util */
#define S1(x) #x
#define S2(x) S1(x)
#define LOCATION __FILE__ ":" S2(__LINE__)

/* Logger helpers */
#define log_info(msg) or_loginfo(msg, LOCATION)
#define log_warn(msg) or_logwarn(msg, LOCATION)
#define log_error(msg) or_logerror(msg, LOCATION)
#define log_fatal(msg) or_logfatal(msg, LOCATION)

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

typedef enum {
    OR_METHOD_UNKNOWN = 0,
    OR_METHOD_GET = 1 << 1,
    OR_METHOD_HEAD = 1 << 2,
    OR_METHOD_POST = 1 << 3,
    OR_METHOD_PUT = 1 << 4,
    OR_METHOD_DELETE = 1 << 5,
    OR_METHOD_PATCH = 1 << 6,
    OR_METHOD_OPTIONS = 1 << 7,
    OR_METHOD_ANY = ~((uint8_t) 0)
} or_method_t;

typedef uint64_t muid_t;

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
    muid_t muid;
    void (*loginfo)(char* msg, char* module_);
    void (*logwarn)(char* msg, char* module_);
    void (*logerror)(char* msg, char* module_);
    void (*logfatal)(char* msg, char* module_);
    uint64_t (*register_http)(muid_t muid, or_method_t method_mask, char* path, or_http_handler_t handler, void* extra);
    uint64_t (*unregister_http)(muid_t muid, or_method_t method_mask, char* path);
} or_api_t;

typedef struct {
    uint32_t version;
} health_struct_t;

typedef bool (*init_func_t)(const or_api_t* api);
typedef bool (*uninit_func_t)(const or_api_t* api);

/* Exported functions from modmgr.go */
extern uint64_t or_register_http(muid_t muid, or_method_t method_mask, char* path, or_http_handler_t handler, void* extra);
extern uint64_t or_unregister_http(muid_t muid, or_method_t method_mask, char* path);

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
mod_handle_t cffi_load_module(char* path, muid_t muid);
void cffi_unload_module(mod_handle_t handle, muid_t muid);
void call_or_http_handler(or_http_handler_t fn, or_ctx_t* ctx, or_http_req_t* req, void* extra);
loadmod_err_t get_error(void);

#endif // CFFI_H
