package models

import "k8s.io/apimachinery/pkg/types"

type CommittedRun struct {
	JobUID    types.UID
	Algorithm string
	RequestId string
}
