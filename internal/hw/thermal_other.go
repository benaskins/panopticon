//go:build !darwin

package hw

func ThermalState() string { return "Unknown" }
