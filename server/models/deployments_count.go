package models

// DeploymentsCount has the vms and ips reserved in the grid
type DeploymentsCount struct {
	VMs int64 `json:"vms"`
	IPs int64 `json:"ips"`
}

// CountAllDeployments returns deployments and IPs count
func (d *DB) CountAllDeployments() (DeploymentsCount, error) {
	var vmsCount int64
	result := d.db.Table("vms").Count(&vmsCount)
	if result.Error != nil {
		return DeploymentsCount{}, result.Error
	}

	var k8sCount int64
	result = d.db.Table("masters").Count(&k8sCount)
	if result.Error != nil {
		return DeploymentsCount{}, result.Error
	}

	dlsCount := k8sCount + vmsCount

	var vmIPsCount int64
	result = d.db.Table("vms").Where("public_ip = true").Count(&vmIPsCount)
	if result.Error != nil {
		return DeploymentsCount{}, result.Error
	}

	var k8sIPsCount int64
	result = d.db.Table("masters").Where("public_ip = true").Count(&k8sIPsCount)
	if result.Error != nil {
		return DeploymentsCount{}, result.Error
	}

	ipsCount := k8sIPsCount + vmIPsCount

	return DeploymentsCount{
		dlsCount, ipsCount,
	}, result.Error
}
