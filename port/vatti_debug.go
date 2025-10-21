package clipper

import (
	"fmt"
	"io"
	"os"
)

// Debug logging infrastructure for Vatti algorithm
var (
	// VattiDebug enables detailed debug logging when true
	VattiDebug = false
	// VattiDebugOutput is where debug output goes (default: os.Stdout)
	VattiDebugOutput io.Writer = os.Stdout
)

// debugLog prints a debug message if VattiDebug is enabled
func debugLog(format string, args ...interface{}) {
	if VattiDebug {
		fmt.Fprintf(VattiDebugOutput, "[VATTI] "+format+"\n", args...)
	}
}

// debugLogPhase prints a phase separator in debug output
func debugLogPhase(phase string) {
	if VattiDebug {
		fmt.Fprintf(VattiDebugOutput, "\n========================================\n")
		fmt.Fprintf(VattiDebugOutput, "PHASE: %s\n", phase)
		fmt.Fprintf(VattiDebugOutput, "========================================\n\n")
	}
}

// debugLogEdge prints detailed edge information
func debugLogEdge(label string, e *Edge) {
	if VattiDebug && e != nil {
		fmt.Fprintf(VattiDebugOutput, "  %s:\n", label)
		fmt.Fprintf(VattiDebugOutput, "    Bot: %v, Top: %v\n", e.Bot, e.Top)
		fmt.Fprintf(VattiDebugOutput, "    CurrX: %d, Dx: %.4f\n", e.CurrX, e.Dx)
		fmt.Fprintf(VattiDebugOutput, "    WindDx: %d, WindCount: %d, WindCount2: %d\n", e.WindDx, e.WindCount, e.WindCount2)
		fmt.Fprintf(VattiDebugOutput, "    IsLeftBound: %v\n", e.IsLeftBound)
		if e.LocalMin != nil {
			fmt.Fprintf(VattiDebugOutput, "    PathType: %v\n", e.LocalMin.PathType)
		}
	}
}

// debugLogAEL prints the entire active edge list
func debugLogAEL(ael *Edge) {
	if !VattiDebug {
		return
	}

	fmt.Fprintf(VattiDebugOutput, "  Active Edge List (left to right):\n")
	if ael == nil {
		fmt.Fprintf(VattiDebugOutput, "    (empty)\n")
		return
	}

	count := 0
	for e := ael; e != nil; e = e.NextInAEL {
		count++
		pathType := "unknown"
		if e.LocalMin != nil {
			if e.LocalMin.PathType == PathTypeSubject {
				pathType = "subject"
			} else {
				pathType = "clip"
			}
		}
		fmt.Fprintf(VattiDebugOutput, "    [%d] X=%d Dx=%.4f WindDx=%d WC=%d/%d Type=%s Left=%v\n",
			count, e.CurrX, e.Dx, e.WindDx, e.WindCount, e.WindCount2, pathType, e.IsLeftBound)
	}
}

// debugLogOutRec prints output record information
func debugLogOutRec(label string, outRec *OutRec) {
	if !VattiDebug || outRec == nil {
		return
	}

	fmt.Fprintf(VattiDebugOutput, "  %s (OutRec #%d):\n", label, outRec.Idx)

	if outRec.Pts == nil {
		fmt.Fprintf(VattiDebugOutput, "    (no points)\n")
		return
	}

	fmt.Fprintf(VattiDebugOutput, "    Points: ")
	start := outRec.Pts
	current := start
	count := 0
	for {
		fmt.Fprintf(VattiDebugOutput, "%v ", current.Pt)
		current = current.Next
		count++
		if current == start || count > 100 {
			break
		}
	}
	fmt.Fprintf(VattiDebugOutput, "\n    Total points: %d\n", count)
}

// debugLogWindingCalc prints winding count calculation details
func debugLogWindingCalc(e *Edge, isContributing bool) {
	if !VattiDebug || e == nil {
		return
	}

	pathType := "unknown"
	if e.LocalMin != nil {
		if e.LocalMin.PathType == PathTypeSubject {
			pathType = "subject"
		} else {
			pathType = "clip"
		}
	}

	fmt.Fprintf(VattiDebugOutput, "    Edge at X=%d (type=%s): WC=%d/%d, WindDx=%d, Contributing=%v\n",
		e.CurrX, pathType, e.WindCount, e.WindCount2, e.WindDx, isContributing)
}
