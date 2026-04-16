# Go-Template Implementation Roadmap

**Current Status:** 50% Complete (Phases 1-4)
**Last Updated:** 2026-04-16
**Next Phase:** BUG FIX (Phase 09) → TESTING → Phase 05+

---

## IMMEDIATE PRIORITIES (This Week)

### 1. FIX PROJECT STRUCTURE BUG [BLOCKING] 🔴
**Effort:** 1-2 hours
**Status:** CRITICAL

Files generated in nested directories (base/, clean-arch/) instead of project root.

**Solution Options:**
```go
// Option A: Strip first directory in processDir()
relPath := strings.TrimPrefix(path, templateDir+"/")
if strings.Contains(relPath, "/") {
    parts := strings.SplitN(relPath, "/", 2)
    relPath = parts[1]  // Remove template category (base, clean-arch, etc)
}

// Option B: Special handling for base templates
if templateDir == "base" {
    relPath = strings.TrimPrefix(path, templateDir+"/")
}
```

**Test After Fix:**
```bash
go run . init test-project --non-interactive
cd test-project
ls -la  # Should see go.mod, README.md, Makefile in root
```

---

### 2. ADD UNIT TEST SUITE [BLOCKING] 🔴
**Effort:** 4-6 hours
**Status:** REQUIRED FOR CONFIDENCE

**Test Files to Create:**
- `cmd/init_test.go` - Validate project name regex, config creation
- `cmd/version_test.go` - Verify version display
- `internal/config/project_test.go` - Test config defaults
- `internal/generator/funcs_test.go` - Test all template functions
- `internal/generator/generator_test.go` - Test file generation, conditionals
- `internal/prompt/prompt_test.go` - Mock survey interactions

**Coverage Target:** 80%+

**Example Test Structure:**
```go
package cmd

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestProjectNameValidation(t *testing.T) {
    tests := []struct {
        name    string
        valid   bool
    }{
        {"my-service", true},
        {"MyProject", true},
        {"test123", true},
        {"123test", false},
        {"my@project", false},
        {"_test", false},
    }

    for _, tt := range tests {
        result := projectNameRegex.MatchString(tt.name)
        assert.Equal(t, tt.valid, result, tt.name)
    }
}
```

---

## SHORT-TERM ROADMAP (Weeks 2-3)

### 3. COMPLETE PHASE 05: CLEAN ARCHITECTURE [HIGH] 🟠
**Effort:** 5 hours
**After:** Bug fix + tests

**Deliverables:**
- Domain layer (entity, repository interface)
- Usecase layer (CRUD business logic)
- Delivery layer (REST handlers)
- Infrastructure layer (database, logger)
- Database migrations
- Working example: User CRUD

**Specs:** See phase-05-clean-architecture.md

---

### 4. COMPLETE PHASE 06: AUTHENTICATION [HIGH] 🟠
**Effort:** 3 hours
**After:** Phase 05

**Deliverables:**
- JWT middleware & token generation
- OAuth2 support
- Refresh token rotation
- Login/register endpoints

**Specs:** See phase-06-authentication.md

---

### 5. COMPLETE PHASE 07: DOCKER & MONITORING [HIGH] 🟠
**Effort:** 3 hours
**After:** Phase 05

**Deliverables:**
- Dockerfile (multi-stage build)
- docker-compose.yml (app + postgres + prometheus + grafana)
- Prometheus scrape configs
- Grafana dashboards
- Health checks

**Specs:** See phase-07-docker-monitoring.md

---

## MEDIUM-TERM ROADMAP (Week 4)

### 6. COMPLETE PHASE 08: CI/CD & TESTING [HIGH] 🟠
**Effort:** 3 hours
**After:** Phase 07

**Deliverables:**
- GitHub Actions workflow
- GitLab CI pipeline
- Test scaffolding (testify, testcontainers)
- Coverage reporting

**Specs:** See phase-08-cicd-testing.md

---

### 7. COMPLETE PHASE 09: CLI POLISH [HIGH] 🟠
**Effort:** 2 hours
**After:** Phase 08 (or parallel)

**Deliverables:**
- Improved error messages (colored output)
- Progress indicators for generation
- Post-generation validation
- Shell completion script
- Better help text

**ALSO:** Verify bug fix from #1 is working

**Specs:** See phase-09-cli-polish.md

---

### 8. COMPLETE PHASE 10: RELEASE [MEDIUM] 🟡
**Effort:** 2 hours
**After:** Phase 09

**Deliverables:**
- GoReleaser configuration
- GitHub release automation
- Installation docs (.md files)
- Changelog template

**Specs:** See phase-10-release.md

---

## RELEASE MILESTONES

### v0.1.0 (MVP) - Ready for First Release ✅
- **Timeline:** After bug fix + tests (5-8 hours)
- **Includes:** Phases 1-4 complete
- **Features:** CLI generates basic Go projects with configurable options
- **Limitations:** No architecture scaffolding, auth, docker, CI/CD templates
- **Status:** All critical bugs fixed, 80%+ test coverage

### v0.2.0 (Full Scaffolding) - Recommended
- **Timeline:** After phases 5-8 complete (14 hours)
- **Adds:** Clean architecture, authentication, docker, CI/CD templates
- **Status:** Feature-complete for production-ready projects

### v1.0.0 (Stable Release)
- **Timeline:** After phase 10 complete (2 hours)
- **Adds:** Release automation, installation docs
- **Status:** Production-ready with proper release process

---

## IMPLEMENTATION DECISION TREE

```
START
  ↓
[BUG FIX] Fix project structure (1-2 hrs)
  ↓
[TESTS] Add unit test suite (4-6 hrs)
  ↓
Release v0.1.0? ← YES → [RELEASE] Phase 10 (2 hrs)
  │                         ↓
  │                    Announce release
  │
  NO
  ↓
[PHASE 05] Clean Architecture (5 hrs)
  ↓
[PHASE 06] Authentication (3 hrs)
  ↓
[PHASE 07] Docker & Monitoring (3 hrs)
  ↓
[PHASE 08] CI/CD & Testing (3 hrs)
  ↓
[PHASE 09] CLI Polish (2 hrs)
  ↓
Release v0.2.0
  ↓
[PHASE 10] Release Setup (2 hrs)
  ↓
Release v1.0.0
```

---

## FILE MODIFICATIONS GUIDE

### Phase 1-4: No changes needed (COMPLETE) ✅

### Phase 9 (Bug Fix) - Critical Changes
**File:** `internal/generator/generator.go`
**Function:** `processDir()`
**Change:** Strip first-level directory for base templates

```go
// In processDir() method:
relPath := strings.TrimPrefix(path, templateDir+"/")

// FOR BASE TEMPLATES ONLY:
if templateDir == "base" {
    // Remove the 'base/' prefix
    parts := strings.SplitN(relPath, string(filepath.Separator), 1)
    if len(parts) > 1 {
        relPath = parts[1]
    } else {
        relPath = ""  // Skip root
    }
}
```

### Phase 5-10: New files per phase specs
See detailed specs in phase-05-clean-architecture.md through phase-10-release.md

---

## TESTING CHECKLIST

### After Bug Fix:
- [ ] Generated project has files in root (not nested)
- [ ] `make build` works
- [ ] `go mod tidy` succeeds
- [ ] Makefile targets accessible

### After Unit Tests:
- [ ] All test files pass (`go test ./...`)
- [ ] Coverage ≥ 80% (`go test -cover ./...`)
- [ ] CI/CD integration

### After Each Phase:
- [ ] Specs match implementation
- [ ] Generated code compiles
- [ ] No new template issues
- [ ] Conditional rendering still works

---

## BRANCH STRATEGY

Recommended git workflow:

```bash
# Bug fix (high priority)
git checkout -b fix/project-structure-bug
# ... implement fix ...
git commit -m "fix: flatten generated project directory structure"
git push origin fix/project-structure-bug
# → Create PR, merge after testing

# Unit tests
git checkout -b feat/unit-tests
# ... add test files ...
git commit -m "test: add comprehensive unit test coverage (80%+)"
git push origin feat/unit-tests
# → Create PR, merge after all tests pass

# Phases 5+
git checkout -b feat/phase-05-clean-architecture
# ... implement templates ...
git commit -m "feat: add clean architecture templates"
# ... repeat for phases 6-10
```

---

## SUCCESS METRICS

### v0.1 MVP Success:
- ✅ Bug fixed: Generated projects have correct structure
- ✅ Tests passing: 80%+ unit test coverage
- ✅ CLI working: All commands functional
- ✅ Template engine: Conditionals working
- ✅ User satisfaction: Can generate working Go project

### v0.2 Full Success:
- ✅ All phases 1-8 complete
- ✅ Generated projects production-ready
- ✅ Architecture, auth, monitoring, CI/CD included
- ✅ Documentation complete

### v1.0 Stable Success:
- ✅ Release automation working
- ✅ Installation methods documented
- ✅ Community adoption growing
- ✅ Bug-free with proper versioning

---

## BLOCKERS & RISKS

| Blocker | Mitigation | Priority |
|---------|-----------|----------|
| Project structure bug | Implement fix in Phase 9 | CRITICAL |
| Zero test coverage | Add comprehensive tests | CRITICAL |
| Phases 5-8 complexity | Break into smaller PRs | HIGH |
| Dependency version drift | Pin versions explicitly | MEDIUM |
| Documentation gaps | Keep README in sync | MEDIUM |

---

## TEAM HANDOFF CHECKLIST

Before handing to implementation:
- [ ] All phase specs finalized
- [ ] Bug fix approach decided
- [ ] Test coverage requirements clear
- [ ] Git workflow established
- [ ] CI/CD pipeline ready
- [ ] Release plan documented

---

## QUICK COMMANDS

```bash
# Build
make build

# Run tests
go test -v ./...
go test -cover ./...

# Generate project
./bin/go-template init my-test-project --non-interactive
cd my-test-project && ls -la

# Clean
make clean

# Verify fix
# After Phase 9, generated project should have:
# - go.mod, README.md, Makefile in root (not in base/)
```

---

**Document Status:** Ready for implementation team
**Next Action:** Approve bug fix approach, start Phase 9
**Timeline:** 5-8 hours to v0.1.0 production-ready
