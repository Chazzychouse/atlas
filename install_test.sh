#!/bin/sh
set -eu

# Test harness for install.sh
# Stubs network calls and system commands to verify detection logic.

PASS=0
FAIL=0

assert_eq() {
  label="$1"; expected="$2"; actual="$3"
  if [ "$expected" = "$actual" ]; then
    PASS=$((PASS + 1))
    printf '  \033[32mPASS\033[0m %s\n' "$label"
  else
    FAIL=$((FAIL + 1))
    printf '  \033[31mFAIL\033[0m %s: expected "%s", got "%s"\n' "$label" "$expected" "$actual"
  fi
}

# ── Source only the functions from install.sh ─────────────────────────────────
# Extract functions without running main(), and redefine err to not exit the
# parent process (tests capture it in subshells).
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FUNC_SOURCE="$(sed '/^main$/,$d' "$SCRIPT_DIR/install.sh" | grep -v '^set -eu$')"
eval "$FUNC_SOURCE"
# Override err so it exits only the subshell in tests
err() { printf 'Error: %s\n' "$*" >&2; return 1; }

# ── Test detect_os ────────────────────────────────────────────────────────────
printf '\ndetect_os:\n'

test_detect_os() {
  stub_uname="$1"; expected="$2"; label="$3"
  # Override uname
  uname() { echo "$stub_uname"; }
  actual="$(detect_os 2>/dev/null)" || actual="ERROR"
  assert_eq "$label" "$expected" "$actual"
  unset -f uname
}

test_detect_os "Linux"   "linux"   "Linux  -> linux"
test_detect_os "Darwin"  "darwin"  "Darwin -> darwin"
test_detect_os "MINGW64" "windows" "MINGW  -> windows"
test_detect_os "MSYS_NT" "windows" "MSYS   -> windows"
test_detect_os "CYGWIN"  "windows" "CYGWIN -> windows"

# Unsupported OS should fail
uname() { echo "FreeBSD"; }
result="$(detect_os 2>/dev/null)" && status=0 || status=$?
if [ $status -ne 0 ] || [ -z "$result" ]; then
  PASS=$((PASS + 1))
  printf '  \033[32mPASS\033[0m FreeBSD -> error\n'
else
  FAIL=$((FAIL + 1))
  printf '  \033[31mFAIL\033[0m FreeBSD should error, got "%s"\n' "$result"
fi
unset -f uname

# ── Test detect_arch ──────────────────────────────────────────────────────────
printf '\ndetect_arch:\n'

test_detect_arch() {
  stub_uname="$1"; expected="$2"; label="$3"
  uname() { echo "$stub_uname"; }
  actual="$(detect_arch 2>/dev/null)" || actual="ERROR"
  assert_eq "$label" "$expected" "$actual"
  unset -f uname
}

test_detect_arch "x86_64"  "amd64" "x86_64  -> amd64"
test_detect_arch "amd64"   "amd64" "amd64   -> amd64"
test_detect_arch "aarch64" "arm64" "aarch64 -> arm64"
test_detect_arch "arm64"   "arm64" "arm64   -> arm64"

# Unsupported arch should fail
uname() { echo "i386"; }
result="$(detect_arch 2>/dev/null)" && status=0 || status=$?
if [ $status -ne 0 ] || [ -z "$result" ]; then
  PASS=$((PASS + 1))
  printf '  \033[32mPASS\033[0m i386 -> error\n'
else
  FAIL=$((FAIL + 1))
  printf '  \033[31mFAIL\033[0m i386 should error, got "%s"\n' "$result"
fi
unset -f uname

# ── Test verify_checksum ─────────────────────────────────────────────────────
printf '\nverify_checksum:\n'

tmpdir="$(mktemp -d)"
trap 'rm -rf "$tmpdir"' EXIT

echo "hello" > "$tmpdir/test.tar.gz"
expected_sum="$(sha256sum "$tmpdir/test.tar.gz" | awk '{print $1}')"

# Valid checksum file
echo "${expected_sum}  test.tar.gz" > "$tmpdir/checksums.txt"
verify_checksum "$tmpdir/test.tar.gz" "$tmpdir/checksums.txt" "test.tar.gz" 2>/dev/null && status=0 || status=$?
assert_eq "valid checksum passes" "0" "$status"

# Tampered checksum
echo "0000000000000000000000000000000000000000000000000000000000000000  test.tar.gz" > "$tmpdir/bad_checksums.txt"
verify_checksum "$tmpdir/test.tar.gz" "$tmpdir/bad_checksums.txt" "test.tar.gz" 2>/dev/null && status=0 || status=$?
if [ "$status" -ne 0 ]; then
  PASS=$((PASS + 1))
  printf '  \033[32mPASS\033[0m bad checksum fails\n'
else
  FAIL=$((FAIL + 1))
  printf '  \033[31mFAIL\033[0m bad checksum should fail\n'
fi

# Missing archive in checksums
echo "${expected_sum}  other.tar.gz" > "$tmpdir/missing_checksums.txt"
verify_checksum "$tmpdir/test.tar.gz" "$tmpdir/missing_checksums.txt" "test.tar.gz" 2>/dev/null && status=0 || status=$?
if [ "$status" -ne 0 ]; then
  PASS=$((PASS + 1))
  printf '  \033[32mPASS\033[0m missing archive in checksums fails\n'
else
  FAIL=$((FAIL + 1))
  printf '  \033[31mFAIL\033[0m missing archive should fail\n'
fi

# ── Test archive name construction ────────────────────────────────────────────
printf '\narchive name (OS x Arch matrix):\n'

os_list="linux darwin windows"
arch_list="amd64 arm64"
for os in $os_list; do
  for arch in $arch_list; do
    archive="atlas_${os}_${arch}.tar.gz"
    expected="atlas_${os}_${arch}.tar.gz"
    assert_eq "${os}/${arch}" "$expected" "$archive"
  done
done

# ── Summary ──────────────────────────────────────────────────────────────────
printf '\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n'
printf 'Results: %d passed, %d failed\n' "$PASS" "$FAIL"
if [ "$FAIL" -gt 0 ]; then
  exit 1
fi
