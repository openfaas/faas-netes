package types

type ScaleServiceRequest struct {
	ServiceName string `json:"serviceName"`
	Replicas    uint64 `json:"replicas"`
}
