package main

// ServiceEndpoints todo
type ServiceEndpoints struct {
	Targets []string `json:"targets"`
}

// Config config
type Config struct {
	OutputPath  string       `json:"outputPath"`
	SrvAddr     string       `json:"srvAddr"`
	SyncTargets []SyncTarget `json:"syncTargets"`
}

// ServicePort service port pair
type ServicePort struct {
	Service  string `json:"service"`
	PortName string `json:"portName"`
}

// SyncTarget sync target
type SyncTarget struct {
	KubeConfigPath string        `json:"kubeConfigPath"`
	Cluster        string        `json:"cluster"`
	Namespace      string        `json:"namespace"`
	Services       []ServicePort `json:"services"`
}
