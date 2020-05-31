package framework

type Clusters interface {
	GetCluster(name string) Cluster
	Cluster
}

type Cluster interface {
	DeployKuma() error
	VerifyKuma() error
	GetKumaCPLogs() (string, error)
	DeleteKuma() error
}
