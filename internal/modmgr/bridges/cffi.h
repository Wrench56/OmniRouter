#ifndef CFFI_H
#define CFFI_H

#include <stdint.h>
#include <stdbool.h>

#define MODLOADER_VERSION 2
#define MAX_VERSION_LENGTH 20

/* Exported functions from logger_cffi.go */
/* Do NOT use these directly. Use logger helper instead */
void loginfo(char* msg, char* module_);
void logwarn(char* msg, char* module_);
void logerror(char* msg, char* module_);
void logfatal(char* msg, char* module_);

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
    LOADMOD_NO_VALID_INIT_FUNC,
    LOADMOD_INIT_STRUCT_NULL
} loadmod_err_t;

typedef struct {
    uint32_t version;
} init_struct_t;

typedef init_struct_t* (*init_func_t)();

#ifdef __linux__
    #include <dlfcn.h>
    #include <alloca.h>
#endif

/* cffi.c exports */
bool health(void);
loadmod_err_t load_module(char* path);

#endif // CFFI_H
