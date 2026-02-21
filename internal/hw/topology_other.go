//go:build !darwin

package hw

func detectTopology() Topology {
	return Topology{}
}
