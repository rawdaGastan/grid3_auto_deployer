// Package deployer for handling deployments
package deployer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"sync"

	"github.com/codescalers/cloud4students/models"
	"github.com/codescalers/cloud4students/streams"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

// ConsumeVMRequest to consume api requests of vm deployments
func (d *Deployer) ConsumeVMRequest(ctx context.Context, pending bool) {
	result, err := d.Redis.Read(streams.ReqVMStreamName, streams.ReqVMConsumerGroupName, 0, pending)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return
		}
		log.Error().Err(err).Msg("failed to read vm stream request")
		return
	}

	var vmWG sync.WaitGroup

	for _, s := range result {
		for _, message := range s.Messages {
			vmWG.Add(1)
			go func(message redis.XMessage) {
				defer vmWG.Done()

				var codeErr int
				var resErr error
				var req streams.VMDeployRequest

				for _, v := range message.Values {
					err = json.Unmarshal([]byte(v.(string)), &req)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal vm request")
						continue
					}

					codeErr, resErr = d.deployVMRequest(ctx, req.User, req.Input, req.AdminSSHKey)
					if resErr != nil {
						log.Error().Err(resErr).Msg("failed to deploy vm request")
						continue
					}
				}

				if err := d.Redis.DB.XAck(streams.ReqVMStreamName, streams.ReqVMConsumerGroupName, message.ID).Err(); err != nil {
					log.Error().Err(err).Msgf("failed to acknowledge vm request with ID: %s", message.ID)
					resErr = err
					codeErr = http.StatusInternalServerError
				}

				msg := fmt.Sprintf("Your virtual machine '%s' failed to be deployed with error: %s", req.Input.Name, resErr)
				if codeErr == 0 {
					msg = fmt.Sprintf("Your virtual machine '%s' is deployed successfully 🎆", req.Input.Name)
				}

				notification := models.Notification{
					UserID: req.User.ID.String(),
					Msg:    msg,
					Type:   models.VMsType,
				}
				err = d.db.CreateNotification(&notification)
				if err != nil {
					log.Error().Err(err).Msgf("failed to create notification: %+v", notification)
				}

			}(message)
		}
		vmWG.Wait()
	}
}

// ConsumeK8sRequest to consume api requests of k8s deployments
func (d *Deployer) ConsumeK8sRequest(ctx context.Context, pending bool) {
	result, err := d.Redis.Read(streams.ReqK8sStreamName, streams.ReqK8sConsumerGroupName, 0, pending)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return
		}
		log.Error().Err(err).Msg("failed to read k8s stream request")
		return
	}

	var k8sWG sync.WaitGroup

	for _, s := range result {
		for _, message := range s.Messages {
			k8sWG.Add(1)
			go func(message redis.XMessage) {
				defer k8sWG.Done()

				var codeErr int
				var resErr error
				var req streams.K8sDeployRequest

				for _, v := range message.Values {
					err = json.Unmarshal([]byte(v.(string)), &req)
					if err != nil {
						log.Error().Err(err).Msg("failed to unmarshal k8s request")
						continue
					}

					codeErr, resErr = d.deployK8sRequest(ctx, req.User, req.Input, req.AdminSSHKey)
					if resErr != nil {
						log.Error().Err(resErr).Msg("failed to deploy k8s request")
						continue
					}
				}

				if err := d.Redis.DB.XAck(streams.ReqK8sStreamName, streams.ReqK8sConsumerGroupName, message.ID).Err(); err != nil {
					log.Error().Err(err).Msgf("failed to acknowledge k8s request with ID: %s", message.ID)
					resErr = err
					codeErr = http.StatusInternalServerError
				}

				msg := fmt.Sprintf("Your kubernetes cluster '%s' failed to be deployed with error: %s", req.Input.MasterName, resErr)
				if codeErr == 0 {
					msg = fmt.Sprintf("Your kubernetes cluster '%s' is deployed successfully 🎆", req.Input.MasterName)
				}

				notification := models.Notification{
					UserID: req.User.ID.String(),
					Msg:    msg,
					Type:   models.K8sType,
				}
				err = d.db.CreateNotification(&notification)
				if err != nil {
					log.Error().Err(err).Msgf("failed to create notification: %+v", notification)
				}

			}(message)
		}
		k8sWG.Wait()
	}
}

func (d *Deployer) consumeVMs() (nets []workloads.Network, vms []*workloads.Deployment, err error) {
	result, err := d.Redis.Read(streams.DeployVMStreamName, streams.DeployVMConsumerGroupName, 5, false)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nets, vms, nil
		}
		return nets, vms, errors.Wrap(err, "failed to read vm stream deployment")
	}

	for _, s := range result {
		for i, message := range s.Messages {
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
				nets = append(nets, vm.Net)
			}

			if err = d.Redis.DB.XAck(streams.DeployVMStreamName, streams.DeployVMConsumerGroupName, s.Messages[i].ID).Err(); err != nil {
				log.Error().Err(err).Msgf("failed to acknowledge vm request with ID: %s", s.Messages[i].ID)
			}
		}
	}

	return
}

func (d *Deployer) consumeK8s() (nets []workloads.Network, clusters []*workloads.K8sCluster, err error) {
	result, err := d.Redis.Read(streams.DeployK8sStreamName, streams.DeployK8sConsumerGroupName, 5, false)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nets, clusters, nil
		}
		return nets, clusters, errors.Wrap(err, "failed to read clusters stream deployment")
	}

	for _, s := range result {
		for i, message := range s.Messages {
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
				nets = append(nets, k8s.Net)
			}

			if err = d.Redis.DB.XAck(streams.DeployK8sStreamName, streams.DeployK8sConsumerGroupName, s.Messages[i].ID).Err(); err != nil {
				log.Error().Err(err).Msgf("failed to acknowledge k8s request with ID: %s", s.Messages[i].ID)
			}
		}
	}

	return
}
