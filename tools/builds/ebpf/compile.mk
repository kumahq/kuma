CC=clang
CFLAGS=-O2 -g  -Wall -target bpf -I/usr/include/$(shell uname -m)-linux-gnu

MACROS :=
DEBUG ?=

MESH_MODE ?= istio

ifeq ($(MESH_MODE),istio)
    MACROS:= $(MACROS) -DMESH=1
else ifeq ($(MESH_MODE),linkerd)
    MACROS:= $(MACROS) -DMESH=2
else ifeq ($(MESH_MODE),kuma)
    MACROS:= $(MACROS) -DMESH=3
else
$(error MESH_MODE $(MESH_MODE) isn't supported)
endif

# see https://stackoverflow.com/questions/15063298/how-to-check-kernel-version-in-makefile
KVER = $(shell uname -r)
KMAJ = $(shell echo $(KVER) | \
sed -e 's/^\([0-9][0-9]*\)\.[0-9][0-9]*\.[0-9][0-9]*.*/\1/')
KMIN = $(shell echo $(KVER) | \
sed -e 's/^[0-9][0-9]*\.\([0-9][0-9]*\)\.[0-9][0-9]*.*/\1/')
KREV = $(shell echo $(KVER) | \
sed -e 's/^[0-9][0-9]*\.[0-9][0-9]*\.\([0-9][0-9]*\).*/\1/')

kver_ge = $(shell \
echo test | awk '{if($(KMAJ) < $(1)) {print 0} else { \
if($(KMAJ) > $(1)) {print 1} else { \
if($(KMIN) < $(2)) {print 0} else { \
if($(KMIN) > $(2)) {print 1} else { \
if($(KREV) < $(3)) {print 0} else { print 1 } \
}}}}}' \
)

# See https://nakryiko.com/posts/bpf-tips-printk/, kernel will auto print newline if version greater than 5.9.0
ifneq ($(call kver_ge,5,8,999),1)
MACROS:= $(MACROS) -DPRINTNL # kernel version less
endif

ifeq ($(DEBUG),1)
    MACROS:= $(MACROS) -DDEBUG
endif

ifeq ($(USE_RECONNECT),1)
    MACROS:= $(MACROS) -DUSE_RECONNECT
endif

TARGETS=mb_connect.o mb_get_sockopts.o mb_redir.o mb_sockops.o mb_bind.o mb_sendmsg.o mb_recvmsg.o mb_tc.o

%.o: %.c
	$(CC) $(CFLAGS) $(MACROS) -c $< -o $@

generate-compilation-database:
	CC="$(CC)" CFLAGS="$(CFLAGS)" MACROS="$(MACROS)" scripts/generate-compilation-database.sh | tee compile_commands.json

compile: $(TARGETS)
