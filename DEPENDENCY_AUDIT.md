# Dependency Audit Report

**Date:** 2026-01-20
**Project:** go.tempura.ink/co
**Go Version:** 1.24.0

## Executive Summary

This audit identifies **outdated packages**, **potential security concerns**, and **unnecessary dependencies** in the project. Key findings include the ability to remove `golang.org/x/exp` dependency entirely by migrating to standard library equivalents, and several significantly outdated transitive dependencies.

---

## 1. Outdated Packages

### Critical Updates Available

| Package | Current Version | Latest Version | Impact |
|---------|----------------|----------------|---------|
| `github.com/panjf2000/ants/v2` | v2.5.0 | v2.11.4 | **Major** - 6 versions behind |
| `golang.org/x/exp` | 2024-12-19 | 2026-01-12 | **Moderate** - Can be replaced with stdlib |
| `github.com/smarty/assertions` | v1.15.0 | v1.16.0 | **Minor** - Test dependency |
| `github.com/spf13/cobra` | v1.2.1 | v1.10.2 | **Major** - Transitive, 8 versions behind |
| `github.com/spf13/pflag` | v1.0.5 | v1.0.10 | **Minor** - Transitive |
| `github.com/sirupsen/logrus` | v1.8.1 | v1.9.4 | **Moderate** - Transitive |
| `golang.org/x/crypto` | 2021-07-11 | 2025-07-XX | **Critical** - 4+ years old, security risk |
| `golang.org/x/sys` | v0.6.0 | v0.40.0 | **Major** - Transitive |
| `golang.org/x/tools` | v0.40.0 | v0.41.0 | **Minor** - Transitive |
| `golang.org/x/mod` | v0.31.0 | v0.32.0 | **Minor** - Transitive |

### Transitive Dependencies

Many outdated packages are pulled in by `github.com/gopherjs/gopherjs` (via `goconvey`), which itself has not been updated since 2022. The `gopherjs` dependency chain includes:
- Very old versions of `golang.org/x/crypto` (July 2021) - **SECURITY CONCERN**
- Old versions of `golang.org/x/sys`, `golang.org/x/tools`, `github.com/spf13/cobra`

---

## 2. Security Vulnerabilities

### High Priority

**⚠️ `golang.org/x/crypto` v0.0.0-20210711020723-a769d52b0f97**
- **Age:** 4.5+ years old (July 2021)
- **Risk:** Known CVEs likely present in such an old crypto library
- **Source:** Transitive dependency via `gopherjs/gopherjs`
- **Recommendation:** Cannot update directly; comes from goconvey's gopherjs dependency

### Medium Priority

**Old transitive dependencies**
- Multiple golang.org/x packages are significantly outdated
- These are all transitive through `goconvey` and its `gopherjs` dependency

### Note on Vulnerability Scanning

Automated vulnerability scanning with `govulncheck` failed due to network restrictions. Manual inspection reveals the age-based risk above. Consider running `govulncheck ./...` in an environment with internet access to check against the Go vulnerability database.

---

## 3. Unnecessary Dependencies & Bloat

### Can Be Removed Entirely

**`golang.org/x/exp` - HIGH PRIORITY REMOVAL**

Since Go 1.21, functionality from `golang.org/x/exp` has graduated to the standard library:
- `golang.org/x/exp/constraints` → `cmp` (stdlib)
- `golang.org/x/exp/slices` → `slices` (stdlib)

**Current Usage:**
```go
// utils.go:4
import "golang.org/x/exp/constraints"

// Used in utils.go:19
func EvertGET[T constraints.Ordered](ele []T, target T) bool

// Test files use golang.org/x/exp/slices
// async_combine_latest_sequence_test.go
// async_any_sequence_test.go
// ds/queue/multi_receiver_queue_test.go
```

**Migration:** Replace with stdlib equivalents:
- `constraints.Ordered` → `cmp.Ordered`
- `slices` package → `slices` (stdlib)

**Impact:** Removes entire dependency, reduces module size, improves build times

### Benchmark-Only Dependencies

**`github.com/Jeffail/tunny` v0.1.4**
- **Status:** Only used in benchmark comparisons
- **Last Update:** 2019 (7+ years old)
- **Usage:** benchmark/pool_fib_test.go, benchmark/pool_large_json_test.go
- **Recommendation:** Keep for benchmarking, but consider documenting as benchmark-only

**`github.com/panjf2000/ants/v2` v2.5.0**
- **Status:** Only used in benchmark comparisons
- **Current:** v2.5.0, **Latest:** v2.11.4 (6 versions behind)
- **Usage:** Same benchmark files
- **Recommendation:** Update to v2.11.4 OR remove if benchmarks are not regularly run

### Indirect Test Dependencies

**`github.com/gopherjs/gopherjs` v1.17.2**
- **Purpose:** JavaScript transpiler (not needed for this library)
- **Why present:** Transitive dependency of `github.com/smartystreets/goconvey`
- **Bloat:** Pulls in cobra, pflag, crypto, and many other heavy dependencies
- **Size Impact:** Significant (entire Go-to-JS compiler toolchain)
- **Recommendation:** Consider switching test framework to eliminate this bloat

---

## 4. Dependency Tree Analysis

### Main Module Dependencies

```
go.tempura.ink/co
├── github.com/smartystreets/goconvey v1.8.1  [TEST ONLY]
│   └── github.com/gopherjs/gopherjs v1.17.2  [BLOAT]
│       ├── github.com/spf13/cobra v1.2.1  [OUTDATED]
│       ├── golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97  [SECURITY RISK]
│       └── [many more transitive deps]
└── golang.org/x/exp v0.0.0-20251219203646-944ab1f22d93  [CAN BE REMOVED]
```

### Benchmark Module Dependencies

```
benchmark/
├── github.com/Jeffail/tunny v0.1.4  [OUTDATED, BENCHMARK ONLY]
├── github.com/panjf2000/ants/v2 v2.5.0  [OUTDATED, BENCHMARK ONLY]
└── go.tempura.ink/co (local replace)
```

---

## 5. Recommendations

### Immediate Actions (High Priority)

1. **Remove `golang.org/x/exp` dependency**
   - Replace `golang.org/x/exp/constraints.Ordered` with `cmp.Ordered` in utils.go
   - Replace `golang.org/x/exp/slices` with stdlib `slices` in test files
   - Run tests to verify compatibility
   - **Benefit:** Removes entire dependency, no external imports needed

2. **Update benchmark dependencies**
   - Update `github.com/panjf2000/ants/v2` from v2.5.0 to v2.11.4
   - Consider whether `github.com/Jeffail/tunny` is still needed (last updated 2019)
   - Update benchmark module go version from 1.18 to 1.24

3. **Address benchmark go.mod**
   - Current: `go 1.18`
   - Main module: `go 1.24.0`
   - Update benchmark module to match main module version

### Medium Priority

4. **Consider alternative test framework**
   - Current: `github.com/smartystreets/goconvey` (brings gopherjs + heavy deps)
   - Alternative: Standard library `testing` package (zero dependencies)
   - Alternative: `github.com/stretchr/testify` (lighter weight, more popular)
   - **Benefit:** Eliminates gopherjs and all its outdated transitive dependencies including old crypto library

5. **Update transitive dependencies**
   - If keeping goconvey, open issue with maintainers about outdated gopherjs dependency
   - Consider forking goconvey to update gopherjs if no upstream movement

### Low Priority (Best Practices)

6. **Run vulnerability scanning**
   - Run `govulncheck ./...` in environment with internet access
   - Set up CI/CD pipeline to run vulnerability checks on each commit
   - Consider using GitHub Dependabot for automated dependency updates

7. **Document benchmark dependencies**
   - Add comment in benchmark/go.mod explaining these are comparison-only dependencies
   - Consider moving benchmarks to separate repository if they add maintenance burden

8. **Audit indirect dependencies periodically**
   - Set up quarterly dependency audits
   - Monitor security advisories for Go ecosystem

---

## 6. Migration Guide

### Step 1: Remove golang.org/x/exp

**File: utils.go**
```diff
-import "golang.org/x/exp/constraints"
+import "cmp"

-func EvertGET[T constraints.Ordered](ele []T, target T) bool
+func EvertGET[T cmp.Ordered](ele []T, target T) bool
```

**Files: Test files using slices**
```diff
-import "golang.org/x/exp/slices"
+import "slices"
```

Then run:
```bash
go mod tidy
go test ./...
```

### Step 2: Update Benchmark Dependencies

```bash
cd benchmark
go get github.com/panjf2000/ants/v2@latest
go mod tidy
go test -bench=. ./...
cd ..
```

### Step 3: Update Benchmark Go Version

**File: benchmark/go.mod**
```diff
-go 1.18
+go 1.24
```

### Step 4: Verify Everything Works

```bash
# Main module
go test ./...
go build ./...

# Benchmarks
cd benchmark
go test -bench=. ./...
cd ..
```

---

## 7. Summary

| Category | Count | Severity |
|----------|-------|----------|
| Outdated Direct Dependencies | 2 | Low-Moderate |
| Outdated Transitive Dependencies | 10+ | Moderate-High |
| Security Concerns | 1 | High (old crypto lib) |
| Removable Dependencies | 1 | N/A (golang.org/x/exp) |
| Benchmark-Only Dependencies | 2 | Low |

### Key Metrics

- **Total direct dependencies:** 2 (main), 2 (benchmark)
- **Total transitive dependencies:** 20+
- **Dependencies that can be removed:** 1 (golang.org/x/exp)
- **Dependencies that should be updated:** 10+
- **Security-sensitive outdated deps:** 1 (golang.org/x/crypto via gopherjs)

### Estimated Impact of Recommendations

- **Removing golang.org/x/exp:** -1 dependency, faster builds, no external code
- **Updating benchmark deps:** Better performance comparisons, modern APIs
- **Switching away from goconvey:** -15+ transitive dependencies, remove security risk
- **Overall:** Leaner, more secure, more maintainable codebase

---

## 8. Next Steps

1. Create branch for dependency cleanup
2. Implement Step 1 (remove golang.org/x/exp) - **QUICK WIN**
3. Implement Steps 2-3 (update benchmarks)
4. Test thoroughly
5. Consider goconvey replacement in separate PR
6. Set up automated dependency scanning in CI/CD

