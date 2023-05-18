// Package streams for redis streams
package streams

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

// PushVM pushes a vm deployment to the stream
func (r *RedisClient) PushVM(vm VMDeployment) error {
	bytes, err := json.Marshal(vm)
	if err != nil {
		return err
	}

	return r.DB.XAdd(&redis.XAddArgs{
		Stream: DeployVMStreamName,
		Values: map[string]interface{}{vm.DL.Name: bytes},
	}).Err()
}

// PushK8s pushes a k8s cluster deployment to the stream
func (r *RedisClient) PushK8s(k8s K8sDeployment) error {
	bytes, err := json.Marshal(k8s)
	if err != nil {
		return err
	}

	return r.DB.XAdd(&redis.XAddArgs{
		Stream: DeployK8sStreamName,
		Values: map[string]interface{}{k8s.DL.Master.Name: bytes},
	}).Err()
}

// PushVMRequest pushes a vm request to the stream
func (r *RedisClient) PushVMRequest(vm VMDeployRequest) error {
	bytes, err := json.Marshal(vm)
	if err != nil {
		return err
	}

	return r.DB.XAdd(&redis.XAddArgs{
		Stream: ReqVMStreamName,
		Values: map[string]interface{}{string(bytes): bytes},
	}).Err()
}

// PushK8sRequest pushes a k8s request to the stream
func (r *RedisClient) PushK8sRequest(k8s K8sDeployRequest) error {
	bytes, err := json.Marshal(k8s)
	if err != nil {
		return err
	}

	return r.DB.XAdd(&redis.XAddArgs{
		Stream: ReqK8sStreamName,
		Values: map[string]interface{}{string(bytes): bytes},
	}).Err()
}
