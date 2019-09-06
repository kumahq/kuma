package context

type Context struct {
	ControlPlane *ControlPlaneContext
}

type ControlPlaneContext struct {
	SdsLocation string
	SdsTlsCert  []byte
}
