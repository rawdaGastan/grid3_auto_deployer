// Package routes for API endpoints
package routes

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/codescalers/cloud4students/streams"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

func (r *Router) consumeVMRequest(ctx context.Context) {
	result, err := r.redis.DB.XReadGroup(&redis.XReadGroupArgs{
		Streams: []string{streams.ReqVMStreamName, ">"},
		Group:   streams.ReqVMConsumerGroupName,
		Block:   1 * time.Second,
	}).Result()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return
		}
		log.Error().Err(err).Msg("failed to read vm stream request")
		return
	}

	for _, s := range result {
		r.vmRequested = false
		for _, message := range s.Messages {
			for _, v := range message.Values {
				var req streams.VMDeployRequest
				err = json.Unmarshal([]byte(v.(string)), &req)
				if err != nil {
					log.Error().Err(err).Msg("failed to unmarshal vm request")
				}

				_, err := r.deployVMRequest(ctx, req.User, req.Input)
				if err != nil {
					log.Error().Err(err).Msg("failed to deploy vm request")
					//writeErrResponseWithoutLabels(req.Writer.W, errCode, err.Error())
					continue
				}
			}

			if err := r.redis.DB.XAck(streams.ReqVMStreamName, streams.ReqVMConsumerGroupName, message.ID).Err(); err != nil {
				log.Error().Err(err).Msgf("failed to acknowledge vm request with ID: %s", message.ID)
				continue
			}
			r.vmRequested = true
		}
	}
}

func (r *Router) consumeVMs(ctx context.Context) (vms []*workloads.Deployment, err error) {
	result, err := r.redis.DB.XReadGroup(&redis.XReadGroupArgs{
		Streams: []string{streams.DeployVMStreamName, ">"},
		Group:   streams.DeployVMConsumerGroupName,
		Block:   1 * time.Second,
	}).Result()

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
					return vms, errors.Wrap(err, "failed to unmarshal vm request")
				}
			}

			if !reflect.DeepEqual(vm, streams.VMDeployment{}) {
				vms = append(vms, vm.DL)
			}

			fmt.Printf("s.Messages[i].ID: %v\n", message.ID)
		}
	}

	return

	/*err = r.tfPluginClient.DeploymentDeployer.BatchDeploy(ctx, vms)
	if err != nil {
		log.Error().Err(err).Msg("failed to batch deploy vm")
	}

	if err := r.redis.DB.XAck(streams.ReqVMStreamName, streams.ReqVMConsumerGroupName, s.Messages[i].ID).Err(); err != nil {
		log.Error().Err(err).Msgf("failed to acknowledge vm request with ID: %s", s.Messages[i].ID)
		return
	}
	r.vmDeployed = true*/
}

func (r *Router) consumeNets(ctx context.Context) (nets []*workloads.ZNet, err error) {
	result, err := r.redis.DB.XReadGroup(&redis.XReadGroupArgs{
		Streams: []string{streams.DeployNetStreamName, ">"},
		Group:   streams.DeployNetConsumerGroupName,
		Block:   1 * time.Second,
	}).Result()

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
					return nets, errors.Wrap(err, "failed to unmarshal network request")
				}
			}

			if !reflect.DeepEqual(net, streams.NetDeployment{}) {
				nets = append(nets, net.DL)
			}

			fmt.Printf("network s.Messages[i].ID: %v\n", message.ID)
		}
	}

	return
}
