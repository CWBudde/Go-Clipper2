//go:build clipper_cgo
#include "clipper_bridge.h"

// Include the main clipper header which provides all inline implementations
#include <clipper2/clipper.h>

// Include the actual implementation files to resolve undefined references
#include "../third_party/clipper2/CPP/Clipper2Lib/src/clipper.engine.cpp"
#include "../third_party/clipper2/CPP/Clipper2Lib/src/clipper.offset.cpp"
#include "../third_party/clipper2/CPP/Clipper2Lib/src/clipper.rectclip.cpp"

#include <vector>
#include <new>      // std::nothrow
#include <cstring>  // memcpy

using namespace Clipper2Lib;

// ---- helpers: convert between C and C++ types --------------------------------

static Paths64 toPaths64(const cpaths64* cp) {
  Paths64 out;
  if (!cp || cp->len <= 0) return out;
  out.reserve(cp->len);
  for (int i = 0; i < cp->len; ++i) {
    const cpath64& cpath = cp->items[i];
    Path64 p;
    if (cpath.len > 0 && cpath.pts) {
      p.reserve(cpath.len);
      for (int j = 0; j < cpath.len; ++j) {
        p.push_back(Point64{ cpath.pts[j].x, cpath.pts[j].y });
      }
    }
    out.push_back(std::move(p));
  }
  return out;
}

static bool fromPaths64(const Paths64& src, cpaths64* dst) {
  if (!dst) return false;
  dst->len   = (int)src.size();
  dst->items = (cpath64*)operator new[](sizeof(cpath64) * dst->len, std::nothrow);
  if (dst->len && !dst->items) return false;

  for (int i = 0; i < dst->len; ++i) {
    const Path64& sp = src[(size_t)i];
    cpath64& dp = dst->items[i];
    dp.len = (int)sp.size();
    if (dp.len) {
      dp.pts = (cpt64*)operator new[](sizeof(cpt64) * dp.len, std::nothrow);
      if (!dp.pts) return false;
      for (int j = 0; j < dp.len; ++j) {
        dp.pts[j].x = sp[(size_t)j].x;
        dp.pts[j].y = sp[(size_t)j].y;
      }
    } else {
      dp.pts = nullptr;
    }
  }
  return true;
}

extern "C" {

void clipper2c_free_paths(cpaths64* p) {
  if (!p || !p->items) return;
  for (int i = 0; i < p->len; ++i) {
    if (p->items[i].pts) operator delete[](p->items[i].pts);
  }
  operator delete[](p->items);
  p->items = nullptr;
  p->len = 0;
}

static FillRule toFR(c_fillrule fr) {
  switch (fr) {
    case C_FILL_NONZERO:  return FillRule::NonZero;
    case C_FILL_POSITIVE: return FillRule::Positive;
    case C_FILL_NEGATIVE: return FillRule::Negative;
    case C_FILL_EVENODD:
    default:              return FillRule::EvenOdd;
  }
}

static ClipType toCT(c_cliptype ct) {
  switch (ct) {
    case C_CLIP_UNION:        return ClipType::Union;
    case C_CLIP_DIFFERENCE:   return ClipType::Difference;
    case C_CLIP_XOR:          return ClipType::Xor;
    case C_CLIP_INTERSECTION:
    default:                  return ClipType::Intersection;
  }
}

static JoinType toJT(c_jointype jt) {
  switch (jt) {
    case C_JOIN_SQUARE: return JoinType::Square;
    case C_JOIN_BEVEL:  return JoinType::Bevel;
    case C_JOIN_ROUND:  return JoinType::Round;
    case C_JOIN_MITER:  return JoinType::Miter;
    default:            return JoinType::Square;
  }
}

static EndType toET(c_endtype et) {
  switch (et) {
    case C_END_POLYGON: return EndType::Polygon;
    case C_END_JOINED:  return EndType::Joined;
    case C_END_BUTT:    return EndType::Butt;
    case C_END_SQUARE:  return EndType::Square;
    case C_END_ROUND:   return EndType::Round;
    default:            return EndType::Polygon;
  }
}

int clipper2c_boolean64(c_cliptype ct, c_fillrule fr,
                        const cpaths64* subjects,
                        const cpaths64* subjects_open,
                        const cpaths64* clips,
                        cpaths64* out_closed,
                        cpaths64* out_open) {
  try {
    Paths64 subj   = toPaths64(subjects);
    Paths64 subjOp = toPaths64(subjects_open);
    Paths64 clip   = toPaths64(clips);

    Paths64 sol, solOpen;
    
    // Always use Clipper64 when we have open subjects or need to handle both closed and open paths
    Clipper64 clipper;
    bool useClipper = !subjOp.empty() || (!subj.empty() && !clip.empty());
    
    if (useClipper) {
      if (!subj.empty()) {
        clipper.AddSubject(subj);
      }
      if (!subjOp.empty()) {
        clipper.AddOpenSubject(subjOp);
      }
      if (!clip.empty()) {
        clipper.AddClip(clip);
      }
      clipper.Execute(toCT(ct), toFR(fr), sol, solOpen);
    } else if (!subj.empty() && clip.empty()) {
      // Simple case: only closed subjects, no clips - use simpler BooleanOp
      sol = BooleanOp(toCT(ct), toFR(fr), subj, clip);
      // No open paths in this case
    }

    if (out_closed && !fromPaths64(sol, out_closed)) return 2;
    if (out_open   && !fromPaths64(solOpen, out_open)) return 3;
    return 0;
  } catch (...) {
    return 1;
  }
}

int clipper2c_offset64(const cpaths64* paths, double delta,
                       c_jointype jt, c_endtype et,
                       double miter_limit, double arc_tolerance,
                       cpaths64* out_paths) {
  try {
    Paths64 in = toPaths64(paths);
    Paths64 out = InflatePaths(in, delta, toJT(jt), toET(et), miter_limit, arc_tolerance);
    if (!fromPaths64(out, out_paths)) return 2;
    return 0;
  } catch (...) {
    return 1;
  }
}

int clipper2c_rectclip64(int64_t left, int64_t top, int64_t right, int64_t bottom,
                         const cpaths64* in_paths,
                         cpaths64* out_paths) {
  try {
    Paths64 in = toPaths64(in_paths);
    Rect64 r{left, top, right, bottom};
    Paths64 out = RectClip(r, in);
    if (!fromPaths64(out, out_paths)) return 2;
    return 0;
  } catch (...) {
    return 1;
  }
}

} // extern "C"