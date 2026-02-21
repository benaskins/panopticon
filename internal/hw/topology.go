package hw

// Topology describes the static hardware configuration detected at startup.
type Topology struct {
	PCores      int // total P (performance) cores
	ECores      int // total E (efficiency) cores
	PPerCluster int // P-cores per L2 cluster
	EPerCluster int // E-cores per L2 cluster
	PClusters   int // number of P-core clusters
	EClusters   int // number of E-core clusters
	GPUCores    int // total GPU cores
	GPUGroups   int // MGPU group count
	TotalMemory uint64 // total physical RAM in bytes
}

// DetectTopology reads static hardware info. Platform-specific.
func DetectTopology() Topology {
	return detectTopology()
}
