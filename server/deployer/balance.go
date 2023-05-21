// Package deployer for handling deployments
package deployer

import (
	"github.com/pkg/errors"
)

// GetBalance returns the current balance of the deployer account
func (d *Deployer) GetBalance() (float64, error) {
	balance, err := d.tfPluginClient.SubstrateConn.GetBalance(d.tfPluginClient.Identity)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get account balance with the given mnemonics")
	}

	return float64(balance.Free.Int64()) / 1e7, nil
}
