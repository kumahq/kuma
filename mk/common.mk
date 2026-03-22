# Reusable Make functions and variables shared across mk/*.mk files.
# Include-guarded so it can be pulled in by individual mk files or the top-level Makefile.
ifndef _COMMON_MK
_COMMON_MK := 1

# --- Verbosity ---

# Set V=1 on the command line for full recipe output.
V ?=
Q := $(if $(V),,@)

# --- String helpers ---

_digits := 0 1 2 3 4 5 6 7 8 9

# _strip_chars(chars, text): fold over $(1), removing each character from $(2)
_strip_chars = $(if $(1),\
  $(call _strip_chars,\
    $(wordlist 2,$(words $(1)),$(1)),\
    $(subst $(firstword $(1)),,$(2))),\
  $(2))

# _is_digits(x): "yes" iff x is non-empty and contains only digits
_is_digits = $(if $(strip $(1)),$(if $(strip $(call _strip_chars,$(_digits),$(1))),,yes))

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
