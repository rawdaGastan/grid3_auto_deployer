// Package streams for redis streams
package streams

import (
	"encoding/json"

	"github.com/go-redis/redis"
)

// PushNet pushes a network deployment to the stream
func (r *RedisClient) PushNet(net NetDeployment) error {
	bytes, err := json.Marshal(net)
	if err != nil {
		return err
	}

	return r.DB.XAdd(&redis.XAddArgs{
		Stream: DeployNetStreamName,
		Values: map[string]interface{}{net.DL.Name: bytes},
	}).Err()
}

// PushVM pushes a vm deployment to the stream
func (r *RedisClient) PushVM(net NetDeployment, vm VMDeployment) error {
	err := r.PushNet(net)
	if err != nil {
		return err
	}

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
func (r *RedisClient) PushK8s(net NetDeployment, k8s K8sDeployment) error {
	err := r.PushNet(net)
	if err != nil {
		return err
	}

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