CC      ?= cc
CFLAGS  ?= -O2 -fPIC -Wall -Wextra -Wpedantic
LDFLAGS ?=
LDLIBS  ?=

SRC := src/helloworld.c

ifeq ($(OS),Windows_NT)
    TARGET  := build/helloworld.dll
    SOFLAGS := -shared
else
    UNAME_S := $(shell uname -s)
    ifeq ($(UNAME_S),Darwin)
		# TODO: Test this
        TARGET := build/helloworld.dylib
        SOFLAGS := -dynamiclib -Wl,-install_name,@rpath/libhelloworld.dylib
    else
        TARGET := build/helloworld.so
        SOFLAGS := -shared
    endif
endif

all: $(TARGET)

$(TARGET): $(SRC)
	$(CC) $(CFLAGS) $(LDFLAGS) $(SOFLAGS) -o $@ $< $(LDLIBS)

clean:
	@if [ true ]; then \
		$(RM) $(TARGET); \
	fi

.PHONY: clean all
