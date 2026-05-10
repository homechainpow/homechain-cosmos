/*
 * Argon2 reference source code package - reference C implementation
 *
 * Copyright 2015
 * Daniel Dinu, Dmitry Khovratovich, Jean-Philippe Aumasson, and Samuel Neves
 *
 * You may use this work under the terms of a Creative Commons CC0 1.0
 * License/Waiver or the Apache Public License 2.0, at your option. The terms
 * of these licenses can be found at:
 *
 * - CC0 1.0 Universal : http://creativecommons.org/publicdomain/zero/1.0
 * - Apache 2.0        : http://www.apache.org/licenses/LICENSE-2.0
 *
 * You should have received a copy of both of these licenses along with this
 * software. If not, they may be obtained at the above URLs.
 */

#ifndef ARGON2_H
#define ARGON2_H

#include <stddef.h>
#include <stdint.h>

/*
 * Argon2 input type
 */
#define ARGON2_PASSWORD 0
#define ARGON2_SALT 1

/*
 * Argon2 type
 */
#define ARGON2_D 0
#define ARGON2_I 1
#define ARGON2_ID 2

typedef enum Argon2_type {
    Argon2_d = ARGON2_D,
    Argon2_i = ARGON2_I,
    Argon2_id = ARGON2_ID
} argon2_type;

/*
 * Version numbers
 */
#define ARGON2_VERSION_10 0x10
#define ARGON2_VERSION_13 0x13

/*
 * Error codes
 */
typedef enum Argon2_ErrorCodes {
    ARGON2_OK = 0,

    ARGON2_ERROR_MEMORY_ALLOCATION = -1,

    ARGON2_ERROR_INCORRECT_PARAMETER = -2,

    ARGON2_ERROR_OUT_OF_MEMORY = -3,

    ARGON2_ERROR_MEMORY_ALLOCATION_MISMATCH = -4,

    ARGON2_ERROR_THREAD_FAIL = -5,

    ARGON2_ERROR_DECODING_FAIL = -6,

    ARGON2_ERROR_DECODING_LENGTH_FAIL = -7,

    ARGON2_ERROR_MISMATCHING_TYPES = -8,

    ARGON2_ERROR_PWD_PTR_MISMATCH = -9,

    ARGON2_ERROR_SALT_PTR_MISMATCH = -10,

    ARGON2_ERROR_SECRET_PTR_MISMATCH = -11,

    ARGON2_ERROR_AD_PTR_MISMATCH = -12,

    ARGON2_ERROR_OUT_OF_PTR_MISMATCH = -13,

    ARGON2_ERROR_TIME_TOO_SMALL = -14,

    ARGON2_ERROR_TIME_TOO_LARGE = -15,

    ARGON2_ERROR_MEMORY_TOO_LITTLE = -16,

    ARGON2_ERROR_MEMORY_TOO_MUCH = -17,

    ARGON2_ERROR_LANES_TOO_FEW = -18,

    ARGON2_ERROR_LANES_TOO_MANY = -19,

    ARGON2_ERROR_PWD_TOO_SHORT = -20,

    ARGON2_ERROR_PWD_TOO_LONG = -21,

    ARGON2_ERROR_SALT_TOO_SHORT = -22,

    ARGON2_ERROR_SALT_TOO_LONG = -23,

    ARGON2_ERROR_SECRET_TOO_SHORT = -24,

    ARGON2_ERROR_SECRET_TOO_LONG = -25,

    ARGON2_ERROR_AD_TOO_SHORT = -26,

    ARGON2_ERROR_AD_TOO_LONG = -27,

    ARGON2_ERROR_HASH_LEN_MISMATCH = -28,

    ARGON2_ERROR_TIME_MISMATCH = -29,

    ARGON2_ERROR_LANES_MISMATCH = -30,

    ARGON2_ERROR_CPU_COST_MISMATCH = -31,

    ARGON2_ERROR_MEMORY_COST_MISMATCH = -32,

    ARGON2_ERROR_PARALLELISM_MISMATCH = -33,

    ARGON2_ERROR_INCORRECT_TYPE = -34,

    ARGON2_ERROR_OUT_OF_MEMORY_NSEC = -35,

    ARGON2_ERROR_OTHER = -36
} argon2_error_codes;

/* Memory allocator types --- for external allocation */
typedef int (*allocate_fptr)(uint8_t **memory, size_t bytes_to_allocate);
typedef void (*deallocate_fptr)(uint8_t *memory, size_t bytes_to_allocate);

/* --- Argon2 context structure --- */
typedef struct Argon2_Context {
    uint8_t *out;    /* output array */
    uint32_t outlen; /* length of output array */

    uint8_t *pwd;    /* password array */
    uint32_t pwdlen; /* password length */

    uint8_t *salt;    /* salt array */
    uint32_t saltlen; /* salt length */

    uint8_t *secret;    /* key array */
    uint32_t secretlen; /* key length */

    uint8_t *ad;    /* associated data array */
    uint32_t adlen; /* associated data length */

    size_t t_cost;  /* number of passes */
    size_t m_cost;  /* amount of memory requested (KB) */
    size_t lanes;   /* number of parallel lanes */
    size_t threads; /* maximum number of threads */

    uint32_t version; /* version number */

    allocate_fptr allocate_cbk; /* pointer to memory allocator */
    deallocate_fptr free_cbk;   /* pointer to memory deallocator */

    uint32_t flags; /* array of bool flags */
} argon2_context;

/* --- Argon2 external API --- */

int argon2_hash(const uint32_t t_cost, const uint32_t m_cost,
                const uint32_t parallelism, const void *pwd,
                const size_t pwdlen, const void *salt,
                const size_t saltlen, void *hash,
                const size_t hashlen, const char *encoded,
                const size_t encodedlen, argon2_type type,
                const uint32_t version);

const char *argon2_error_message(int error_code);

int argon2d_hash_raw(const uint32_t t_cost, const uint32_t m_cost,
                     const uint32_t parallelism, const void *pwd,
                     const size_t pwdlen, const void *salt,
                     const size_t saltlen, void *hash,
                     const size_t hashlen);

int argon2i_hash_raw(const uint32_t t_cost, const uint32_t m_cost,
                     const uint32_t parallelism, const void *pwd,
                     const size_t pwdlen, const void *salt,
                     const size_t saltlen, void *hash,
                     const size_t hashlen);

int argon2id_hash_raw(const uint32_t t_cost, const uint32_t m_cost,
                      const uint32_t parallelism, const void *pwd,
                      const size_t pwdlen, const void *salt,
                      const size_t saltlen, void *hash,
                      const size_t hashlen);

int argon2_verify(const char *encoded, const void *pwd, const size_t pwdlen,
                  argon2_type type);

int argon2d_verify(const char *encoded, const void *pwd, const size_t pwdlen);

int argon2i_verify(const char *encoded, const void *pwd, const size_t pwdlen);

int argon2id_verify(const char *encoded, const void *pwd, const size_t pwdlen);

int argon2id_hash_encoded(const uint32_t t_cost, const uint32_t m_cost,
                          const uint32_t parallelism, const void *pwd,
                          const size_t pwdlen, const void *salt,
                          const size_t saltlen, const size_t hashlen,
                          char *encoded, const size_t encodedlen);

int argon2id_hash_raw(const uint32_t t_cost, const uint32_t m_cost,
                      const uint32_t parallelism, const void *pwd,
                      const size_t pwdlen, const void *salt,
                      const size_t saltlen, void *hash,
                      const size_t hashlen);

/* --- Argon2 core functions --- */

int argon2_ctx(argon2_context *context, argon2_type type);

#endif
