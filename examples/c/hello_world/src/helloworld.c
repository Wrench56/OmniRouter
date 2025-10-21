#include <stdbool.h>

#include "../../../../internal/modmgr/bridges/cffi.h"

static void (*loginfo_)(char* msg, char* module_);

void hello_world_handler(or_ctx_t* ctx, or_http_req_t* req, void* extra) {
    loginfo_("Hello World triggered!\n", LOCATION);
}

bool init(const or_api_t* api) {
    api->register_http("/", hello_world_handler, NULL);
    api->loginfo("Hello from the dynamically loaded library!\n", LOCATION);

    loginfo_ = api->loginfo;
    return true;
}
