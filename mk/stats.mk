OLD_SHELL := $(SHELL)
SHELL = /bin/bash ${KUMA_DIR}/tools/stats/tracer.sh $(OLD_SHELL) $@ ${KUMA_DIR}
