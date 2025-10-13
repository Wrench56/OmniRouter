#include <stdio.h>
#include <stdlib.h>

#include "../../../../internal/modmgr/bridges/cffi.h"

init_struct_t* init(void) {
    fprintf(stdout, "Hello from the dynamically loaded library!\n");
    fflush(stdout);
    init_struct_t* init_struct = malloc(sizeof(init_struct_t));
    init_struct->version = 1;
    return init_struct;
}
