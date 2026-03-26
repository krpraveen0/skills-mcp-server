#!/usr/bin/env bash
# verify-build.sh — run this from the project root to test both builds locally
set -e

PROJECT_ROOT="$(cd "$(dirname "$0")" && pwd)"
PASS=0
FAIL=0

green() { echo -e "\033[32m✅ $1\033[0m"; }
red()   { echo -e "\033[31m❌ $1\033[0m"; }
blue()  { echo -e "\033[34m▶  $1\033[0m"; }

echo ""
blue "=== Skills MCP Server — Local Build Verification ==="
echo ""

# ── Frontend ──────────────────────────────────────────────────────────────────
blue "[1/5] Frontend: npm install"
cd "$PROJECT_ROOT/frontend"
if npm install --silent; then
  green "npm install succeeded"
  ((PASS++))
else
  red "npm install FAILED"
  ((FAIL++))
fi

blue "[2/5] Frontend: TypeScript type-check (npm run lint)"
if npm run lint; then
  green "TypeScript type-check passed — 0 errors"
  ((PASS++))
else
  red "TypeScript errors found — see output above"
  ((FAIL++))
fi

blue "[3/5] Frontend: Production build (npm run build)"
if npm run build; then
  green "Vite build succeeded — dist/ created"
  ((PASS++))
else
  red "Vite build FAILED"
  ((FAIL++))
fi

# ── Backend ───────────────────────────────────────────────────────────────────
blue "[4/5] Backend: go mod tidy + go build"
cd "$PROJECT_ROOT/backend"
if go mod tidy && go build ./cmd/server/... && go build ./cmd/worker/...; then
  green "Go build succeeded"
  ((PASS++))
else
  red "Go build FAILED"
  ((FAIL++))
fi

blue "[5/5] Backend: go vet"
if go vet ./...; then
  green "go vet passed — no issues"
  ((PASS++))
else
  red "go vet found issues"
  ((FAIL++))
fi

# ── Summary ───────────────────────────────────────────────────────────────────
echo ""
echo "══════════════════════════════════════"
echo "  Results: $PASS passed, $FAIL failed"
echo "══════════════════════════════════════"

if [ "$FAIL" -eq 0 ]; then
  green "All checks passed! Safe to push."
  echo ""
  echo "  Run: bash push-to-github.sh"
else
  red "Fix the failures above before pushing."
  exit 1
fi
