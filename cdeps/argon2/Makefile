# Simple Makefile for Argon2 library
CC = gcc
CFLAGS = -O2 -fPIC
AR = ar
ARFLAGS = rcs

# Directories
INCLUDE_DIR = include
LIB_DIR = lib
SRC_DIR = .

# Files
HEADERS = $(INCLUDE_DIR)/argon2.h
SOURCES = argon2.c
OBJECTS = $(SOURCES:.c=.o)
STATIC_LIB = $(LIB_DIR)/libargon2.a
SHARED_LIB = $(LIB_DIR)/libargon2.dll

# Default target
all: directories $(STATIC_LIB) $(SHARED_LIB)

# Create directories
directories:
	@mkdir -p $(LIB_DIR)

# Static library
$(STATIC_LIB): $(OBJECTS)
	$(AR) $(ARFLAGS) $@ $^

# Shared library (Windows DLL)
$(SHARED_LIB): $(OBJECTS)
	$(CC) -shared -o $@ $^

# Object files
%.o: %.c $(HEADERS)
	$(CC) $(CFLAGS) -I$(INCLUDE_DIR) -c $< -o $@

# Install
install: all
	@echo "Argon2 library built successfully"
	@echo "Static library: $(STATIC_LIB)"
	@echo "Shared library: $(SHARED_LIB)"

# Clean
clean:
	rm -f $(OBJECTS) $(STATIC_LIB) $(SHARED_LIB)
	rm -rf $(LIB_DIR)

.PHONY: all directories install clean
