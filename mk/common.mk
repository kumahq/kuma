# Reusable Make functions and variables shared across mk/*.mk files.
# Include-guarded so it can be pulled in by individual mk files or the top-level Makefile.
ifndef _COMMON_MK
_COMMON_MK := 1

# --- Verbosity ---

# Set V=1 on the command line for full recipe output.
V ?=
Q := $(if $(V),,@)

# --- String helpers ---

# _is_digits(x): "yes" iff x is a single word made only of digits.
# Nested $(subst) instead of recursive $(call), which crashes GNU Make
# (segfault) when evaluated during Makefile parsing in certain versions.
_is_digits = $(if $(filter 1,$(words $(1))),$(if $(subst 0,,$(subst 1,,$(subst 2,,$(subst 3,,$(subst 4,,$(subst 5,,$(subst 6,,$(subst 7,,$(subst 8,,$(subst 9,,$(1))))))))))),,yes))

# --- _retry ---

# Reusable retry wrapper using Make-level loop unrolling.
# $(1) = shell command  $(2) = max attempts  $(3) = log label
# $(foreach) expands at Make time - no shell counter variables or arithmetic.
# $(shell seq ...) generates the exact sequence needed with no hardcoded ceiling.
define _retry
{ \
$(foreach n,$(shell seq 1 $(2)),\
  echo "$(3) (attempt $(n)/$(2))"; \
  $(1) && exit 0; \
  $(if $(filter $(n),$(2)),,sleep 1;) \
) \
  echo "[ERROR] $(3) failed after $(2) attempts" >&2; exit 1; \
}
endef

# --- _flock ---

# Wrap a shell snippet in an exclusive flock when available.
# Uses FD 9 redirection so the lock is held for the entire group.
# On macOS (no flock) the lock step is a no-op - acceptable for local dev.
# $(1) = lock file  $(2) = shell commands to run under lock
define _flock
( \
  if command -v flock >/dev/null 2>&1; then flock -x 9; fi; \
  $(2) \
) 9>"$(1)"
endef

endif # _COMMON_MK
