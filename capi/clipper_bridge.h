//go:build clipper_cgo
#pragma once

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <stdlib.h>

typedef struct { int64_t x, y; } cpt64;

// Simple AoS "paths" layout to keep CGO conversions trivial.
// Allocate with bridge; free with clipper2c_free_paths.
typedef struct {
  int     len;     // number of points
  cpt64*  pts;     // array of points
} cpath64;

typedef struct {
  int      len;    // number of paths
  cpath64* items;  // array of cpath64
} cpaths64;

typedef enum {
  C_CLIP_INTERSECTION = 0,
  C_CLIP_UNION        = 1,
  C_CLIP_DIFFERENCE   = 2,
  C_CLIP_XOR          = 3
} c_cliptype;

typedef enum {
  C_FILL_EVENODD  = 0,
  C_FILL_NONZERO  = 1,
  C_FILL_POSITIVE = 2,
  C_FILL_NEGATIVE = 3
} c_fillrule;

typedef enum {
  C_JOIN_SQUARE = 0,
  C_JOIN_ROUND  = 1,
  C_JOIN_MITER  = 2
} c_jointype;

typedef enum {
  C_END_BUTT   = 0,
  C_END_SQUARE = 1,
  C_END_ROUND  = 2,
  C_END_JOINED = 3
} c_endtype;

// Returns 0 on success; nonzero on failure.
// All out-params are allocated by the bridge; caller must free with clipper2c_free_paths().
int clipper2c_boolean64(c_cliptype ct, c_fillrule fr,
                        const cpaths64* subjects,
                        const cpaths64* subjects_open,
                        const cpaths64* clips,
                        cpaths64* out_closed,
                        cpaths64* out_open);

int clipper2c_offset64(const cpaths64* paths, double delta,
                       c_jointype jt, c_endtype et,
                       double miter_limit, double arc_tolerance,
                       cpaths64* out_paths);

int clipper2c_rectclip64(int64_t left, int64_t top, int64_t right, int64_t bottom,
                         const cpaths64* in_paths,
                         cpaths64* out_paths);

// Frees anything allocated in cpaths64 (deep).
void clipper2c_free_paths(cpaths64* p);

#ifdef __cplusplus
}
#endif