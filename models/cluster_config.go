package models

type ClusterConfig struct {
	Name    string `json:"name"`
	Context string `json:"context"`
}

type Config struct {
	Clusters []ClusterConfig `json:"clusters"`
}
