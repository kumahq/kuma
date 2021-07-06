package nsenter

// NSType defines a namespace type.
type NSType string

// List of namespace types.
// Notice that neither "mnt" nor "user" are listed into this list.
// Because Golang is multithreaded, we get some errors when trying
// to switch to those namespaces, getting "invalid argument".
// The solution is to reexec the current code so that it will call
// into a C constructor, making sure the namespace can be entered
// without multithreading issues.
const (
	NSTypeCGroup NSType = "cgroup"
	NSTypeIPC    NSType = "ipc"
	NSTypeNet    NSType = "net"
	NSTypePID    NSType = "pid"
	NSTypeUTS    NSType = "uts"
)

// CloneFlagsTable is exported so that consumers of this package don't need
// to define this same table again.
var CloneFlagsTable = make(map[NSType]int)

// Namespace describes a namespace that will be entered.
type Namespace struct {
	Path string
	PID  int
	Type NSType
}
