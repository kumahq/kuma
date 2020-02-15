package envoy

type ClusterInfo struct {
	Name   string
	Weight uint32
	Tags   map[string]string
}
