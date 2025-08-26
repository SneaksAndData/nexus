package models

import (
	"fmt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CancellationRequest struct {
	Reason             string `json:"reason"`
	Initiator          string `json:"initiator"`
	CancellationPolicy string `json:"cancellationPolicy,omitempty"`
}

func (r *CancellationRequest) GetPolicy() (*v1.DeletionPropagation, error) {
	defaultPolicy := v1.DeletePropagationBackground
	if r.CancellationPolicy == "" {
		return &defaultPolicy, nil
	}
	policy := v1.DeletionPropagation(r.CancellationPolicy)
	switch policy {
	case v1.DeletePropagationBackground, v1.DeletePropagationOrphan, v1.DeletePropagationForeground:
		return &policy, nil
	default:
		return nil, fmt.Errorf("invalid CancellationPolicy: '%s', accepted values: Orphan, Foreground or Background", r.CancellationPolicy)
	}
}
