// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/codescalers/cloud4students/streams"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

func (r *Router) consumeVMRequest(ctx context.Context, pending bool) {
	result, err := r.redis.Read(streams.ReqVMStreamName, streams.ReqVMConsumerGroupName, pending)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return
		}
		log.Error().Err(err).Msg("failed to read vm stream request")
		return
	}

	for _, s := range result {
		for _, message := range s.Messages {
			r.vmWG.Add(1)
			go func(message redis.XMessage) {
				var codeErr int
				var resErr error
				var req streams.VMDeployRequest

				for _, v := range message.Values {
					err = json.Unmarshal([]byte(v.(string)), &req)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal vm request")
						continue
					}

					codeErr, resErr = r.deployVMRequest(ctx, req.User, req.Input)
					if resErr != nil {
						log.Error().Err(resErr).Msg("failed to deploy vm request")
						continue
					}
				}

				if err := r.redis.DB.XAck(streams.ReqVMStreamName, streams.ReqVMConsumerGroupName, message.ID).Err(); err != nil {
					log.Error().Err(err).Msgf("failed to acknowledge vm request with ID: %s", message.ID)
					resErr = err
					codeErr = http.StatusInternalServerError
				}

				r.mutex.Lock()
				r.vmRequestResponse[fmt.Sprintf("%s %d", req.Input.Name, req.ID)] = streams.ErrResponse{Code: &codeErr, Err: resErr}
				r.mutex.Unlock()

				r.vmWG.Done()
			}(message)
		}
		r.vmWG.Wait()
	}
}

func (r *Router) consumeK8sRequest(ctx context.Context, pending bool) {
	result, err := r.redis.Read(streams.ReqK8sStreamName, streams.ReqK8sConsumerGroupName, pending)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return
		}
		log.Error().Err(err).Msg("failed to read k8s stream request")
		return
	}

	for _, s := range result {
		for _, message := range s.Messages {
			r.k8sWG.Add(1)
			go func(message redis.XMessage) {
				var codeErr int
				var resErr error
				var req streams.K8sDeployRequest

				for _, v := range message.Values {
					err = json.Unmarshal([]byte(v.(string)), &req)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal k8s request")
						continue
					}

					codeErr, resErr = r.deployK8sRequest(ctx, req.User, req.Input)
					if resErr != nil {
						log.Error().Err(resErr).Msg("failed to deploy k8s request")
						continue
					}
				}

				if err := r.redis.DB.XAck(streams.ReqK8sStreamName, streams.ReqK8sConsumerGroupName, message.ID).Err(); err != nil {
					log.Error().Err(err).Msgf("failed to acknowledge k8s request with ID: %s", message.ID)
					resErr = err
					codeErr = http.StatusInternalServerError
				}

				r.mutex.Lock()
				r.k8sRequestResponse[fmt.Sprintf("%s %d", req.Input.MasterName, req.ID)] = streams.ErrResponse{Code: &codeErr, Err: resErr}
				r.mutex.Unlock()

				r.k8sWG.Done()
			}(message)
		}
		r.k8sWG.Wait()
	}
}

func (r *Router) consumeVMs(ctx context.Context) (vms []*workloads.Deployment, err error) {
	result, err := r.redis.Read(streams.DeployVMStreamName, streams.DeployVMConsumerGroupName, false)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return vms, nil
		}
		return vms, errors.Wrap(err, "failed to read vm stream deployment")
	}

	for _, s := range result {
		for i, message := range s.Messages {
			// consume 5 deployments only
			if i == 5 {
				break
			}

			var vm streams.VMDeployment
			for _, v := range message.Values {
				err = json.Unmarshal([]byte(v.(string)), &vm)
				if err != nil {
					log.Err(err).Msg("failed to unmarshal vm request")
					continue
				}
			}

			if !reflect.DeepEqual(vm, streams.VMDeployment{}) {
				vms = append(vms, vm.DL)
			}

			if err = r.redis.DB.XAck(streams.DeployVMStreamName, streams.DeployVMConsumerGroupName, s.Messages[i].ID).Err(); err != nil {
				log.Error().Err(err).Msgf("failed to acknowledge vm request with ID: %s", s.Messages[i].ID)
			}
		}
	}

	return
}

func (r *Router) consumeK8s(ctx context.Context) (clusters []*workloads.K8sCluster, err error) {
	result, err := r.redis.Read(streams.DeployK8sStreamName, streams.DeployK8sConsumerGroupName, false)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return clusters, nil
		}
		return clusters, errors.Wrap(err, "failed to read clusters stream deployment")
	}

	for _, s := range result {
		for i, message := range s.Messages {
			// consume 5 deployments only
			if i == 5 {
				break
			}

			var k8s streams.K8sDeployment
			for _, v := range message.Values {
				err = json.Unmarshal([]byte(v.(string)), &k8s)
				if err != nil {
					log.Err(err).Msg("failed to unmarshal k8s request")
					continue
				}
			}

			if !reflect.DeepEqual(k8s, streams.K8sDeployment{}) {
				clusters = append(clusters, k8s.DL)
			}

			if err = r.redis.DB.XAck(streams.DeployK8sStreamName, streams.DeployK8sConsumerGroupName, s.Messages[i].ID).Err(); err != nil {
				log.Error().Err(err).Msgf("failed to acknowledge k8s request with ID: %s", s.Messages[i].ID)
			}
		}
	}

	return
}

func (r *Router) consumeNets(ctx context.Context) (nets []*workloads.ZNet, err error) {
	result, err := r.redis.Read(streams.DeployNetStreamName, streams.DeployNetConsumerGroupName, false)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nets, nil
		}
		return nets, errors.Wrap(err, "failed to read network stream deployment")
	}

	for _, s := range result {
		for i, message := range s.Messages {
			// consume 10 deployments only
			if i == 10 {
				break
			}

			var net streams.NetDeployment
			for _, v := range message.Values {
				err = json.Unmarshal([]byte(v.(string)), &net)
				if err != nil {
					log.Err(err).Msg("failed to unmarshal network request")
					continue
				}
			}

			if !reflect.DeepEqual(net, streams.NetDeployment{}) {
				nets = append(nets, net.DL)
			}

			if err = r.redis.DB.XAck(streams.DeployNetStreamName, streams.DeployNetConsumerGroupName, s.Messages[i].ID).Err(); err != nil {
				log.Error().Err(err).Msgf("failed to acknowledge k8s request with ID: %s", s.Messages[i].ID)
			}
		}
	}

	return
}
