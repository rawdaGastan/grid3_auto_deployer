// Package models for database models
package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) DB {
	db := NewDB()
	testDir := t.TempDir()

	dbName := "test.db"
	err := db.Connect(testDir + dbName)
	assert.NoError(t, err)
	err = db.Migrate()
	assert.NoError(t, err)
	return db
}

func TestConnect(t *testing.T) {
	db := NewDB()
	testDir := t.TempDir()
	dbName := "test.db"
	t.Run("invalid path", func(t *testing.T) {
		err := db.Connect(testDir + "/another_dir/" + dbName)
		assert.Error(t, err)
	})
	t.Run("valid path", func(t *testing.T) {
		err := db.Connect(testDir + dbName)
		assert.NoError(t, err)
	})
}

func TestCreateUser(t *testing.T) {
	db := setupDB(t)
	err := db.CreateUser(&User{
		Name: "test",
	})
	assert.NoError(t, err)
	var user User
	err = db.db.First(&user).Error
	assert.Equal(t, user.Name, "test")
	assert.NoError(t, err)
}

func TestGetUserByEmail(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found", func(t *testing.T) {
		err := db.CreateUser(&User{
			Name: "test",
		})
		assert.NoError(t, err)
		_, err = db.GetUserByEmail("email")
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("user found", func(t *testing.T) {
		err := db.CreateUser(&User{
			Name:  "test",
			Email: "email",
		})
		assert.NoError(t, err)
		u, err := db.GetUserByEmail("email")
		assert.Equal(t, u.Name, "test")
		assert.Equal(t, u.Email, "email")
		assert.NoError(t, err)
	})
}

func TestGetUserByID(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found", func(t *testing.T) {
		err := db.CreateUser(&User{
			Name: "test",
		})
		assert.NoError(t, err)
		_, err = db.GetUserByID("not-uuid")
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Name:  "test",
			Email: "email",
		}
		err := db.CreateUser(&user)
		assert.NoError(t, err)
		u, err := db.GetUserByID(user.ID.String())
		assert.Equal(t, u.Name, "test")
		assert.Equal(t, u.Email, "email")
		assert.NoError(t, err)
	})
}

func TestUpdatePassword(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found so nothing updated", func(t *testing.T) {
		err := db.UpdatePassword("email", "new-pass")
		assert.NoError(t, err)
		var user User
		err = db.db.First(&user).Error
		assert.Equal(t, err, gorm.ErrRecordNotFound)
		assert.Empty(t, user)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email:          "email",
			HashedPassword: "pass",
		}
		err := db.CreateUser(&user)
		assert.NoError(t, err)
		err = db.UpdatePassword("email", "new-pass")
		assert.NoError(t, err)
		u, err := db.GetUserByEmail("email")
		assert.Equal(t, u.Email, "email")
		assert.Equal(t, u.HashedPassword, "new-pass")
		assert.NoError(t, err)
	})
}

func TestUpdateUserByID(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found so nothing updated", func(t *testing.T) {
		_, err := db.UpdateUserByID("id", User{Email: "email"})
		assert.NoError(t, err)
		var user User
		err = db.db.First(&user).Error
		assert.Equal(t, err, gorm.ErrRecordNotFound)
		assert.Empty(t, user)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email:          "email",
			Voucher:        "voucher",
			HashedPassword: "pass",
		}
		err := db.CreateUser(&user)
		assert.NoError(t, err)
		id, err := db.UpdateUserByID(user.ID.String(), User{Email: "", Voucher: "voucher", HashedPassword: "new-pass", Name: "name"})
		assert.NoError(t, err)
		assert.Equal(t, id, user.ID.String())
		var u User
		err = db.db.First(&u).Error
		// shouldn't change
		assert.Equal(t, u.Email, user.Email)
		assert.Equal(t, u.Voucher, user.Voucher)
		// should change
		assert.Equal(t, u.HashedPassword, "new-pass")
		assert.Equal(t, u.Name, "name")

		assert.NoError(t, err)
	})
}

func TestUpdateVerification(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found so nothing updated", func(t *testing.T) {
		err := db.UpdateVerification("id", true)
		assert.NoError(t, err)
		var user User
		err = db.db.First(&user).Error
		assert.Equal(t, err, gorm.ErrRecordNotFound)
		assert.Empty(t, user)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email: "email",
		}
		err := db.CreateUser(&user)
		assert.Equal(t, user.Verified, false)
		assert.NoError(t, err)
		err = db.UpdateVerification(user.ID.String(), true)
		assert.NoError(t, err)
		var u User
		err = db.db.First(&u).Error
		assert.NoError(t, err)
		assert.Equal(t, u.Verified, true)
	})
}
func TestAddUserVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("user and voucher not found so nothing updated", func(t *testing.T) {
		err := db.AddUserVoucher("id", "voucher")
		assert.NoError(t, err)
		var user User
		var voucher Voucher

		err = db.db.First(&user).Error
		assert.Equal(t, err, gorm.ErrRecordNotFound)
		assert.Empty(t, user)

		err = db.db.First(&voucher).Error
		assert.Equal(t, err, gorm.ErrRecordNotFound)
		assert.Empty(t, voucher)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email: "email",
		}
		voucher := Voucher{
			Voucher: "voucher",
		}
		err := db.CreateUser(&user)
		assert.NoError(t, err)
		err = db.db.Create(&voucher).Error
		assert.NoError(t, err)
		assert.Equal(t, voucher.Used, false)

		err = db.AddUserVoucher(user.ID.String(), "voucher")
		var u User
		var v Voucher
		err = db.db.First(&u).Error
		assert.NoError(t, err)
		assert.Equal(t, u.Voucher, "voucher")

		err = db.db.First(&v).Error
		assert.NoError(t, err)
		assert.Equal(t, v.Used, true)
	})
}

func TestGetNotUsedVoucherByUserID(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.GetNotUsedVoucherByUserID("id")
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher is used", func(t *testing.T) {
		user := User{
			Email:   "email1",
			Voucher: "voucher1",
		}
		err := db.CreateUser(&user)
		assert.NoError(t, err)
		voucher := Voucher{
			UserID:  user.ID.String(),
			Voucher: "voucher1",
			Used:    true,
		}

		err = db.db.Create(&voucher).Error
		assert.NoError(t, err)

		_, err = db.GetNotUsedVoucherByUserID(user.ID.String())
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		user := User{
			Email:   "email2",
			Voucher: "voucher2",
		}
		err := db.CreateUser(&user)
		assert.NoError(t, err)
		voucher := Voucher{
			UserID:  user.ID.String(),
			Voucher: "voucher2",
			Used:    false,
		}

		err = db.db.Create(&voucher).Error
		assert.NoError(t, err)

		v, err := db.GetNotUsedVoucherByUserID(user.ID.String())
		assert.NoError(t, err)
		assert.Equal(t, voucher, v)
	})
}

func TestCreateVM(t *testing.T) {
	db := setupDB(t)
	vm := VM{Name: "vm"}
	err := db.CreateVM(&vm)
	assert.NoError(t, err)
	var v VM
	err = db.db.First(&v).Error
	assert.NoError(t, err)
	assert.Equal(t, v, vm)
}

func TestGetVMByID(t *testing.T) {
	db := setupDB(t)
	t.Run("vm not found", func(t *testing.T) {
		_, err := db.GetVMByID(1)
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("vm found", func(t *testing.T) {
		vm := VM{Name: "vm"}
		err := db.CreateVM(&vm)
		assert.NoError(t, err)

		v, err := db.GetVMByID(vm.ID)
		assert.Equal(t, v, vm)
		assert.NoError(t, err)
	})
}
func TestGetAllVMs(t *testing.T) {
	db := setupDB(t)
	t.Run("no vms with user", func(t *testing.T) {
		vms, err := db.GetAllVms("user")
		assert.NoError(t, err)
		assert.Empty(t, vms)
	})
	t.Run("vms for different users", func(t *testing.T) {
		vm1 := VM{UserID: "user", Name: "vm1"}
		vm2 := VM{UserID: "user", Name: "vm2"}
		vm3 := VM{UserID: "new-user", Name: "vm3"}

		err := db.CreateVM(&vm1)
		assert.NoError(t, err)
		err = db.CreateVM(&vm2)
		assert.NoError(t, err)
		err = db.CreateVM(&vm3)
		assert.NoError(t, err)

		vms, err := db.GetAllVms("user")
		assert.Equal(t, vms, []VM{vm1, vm2})
		assert.NoError(t, err)

		vms, err = db.GetAllVms("new-user")
		assert.Equal(t, vms, []VM{vm3})
		assert.NoError(t, err)
	})
}
func TestDeleteVMByID(t *testing.T) {
	db := setupDB(t)
	t.Run("delete non existing vm", func(t *testing.T) {
		// gorm doesn't return error if vm doesn't exist
		err := db.DeleteVMByID(1)
		assert.NoError(t, err)
	})
	t.Run("delete existing vm", func(t *testing.T) {
		vm := VM{UserID: "user", Name: "vm"}
		err := db.CreateVM(&vm)
		assert.NoError(t, err)

		err = db.DeleteVMByID(vm.ID)
		assert.NoError(t, err)

		var v VM
		err = db.db.First(&v).Error
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
}

func TestDeleteAllVMs(t *testing.T) {
	db := setupDB(t)
	t.Run("delete non existing vms", func(t *testing.T) {
		// gorm doesn't return error if vms don't exist
		err := db.DeleteAllVms("user")
		assert.NoError(t, err)
	})
	t.Run("delete existing vms", func(t *testing.T) {
		vm1 := VM{UserID: "user", Name: "vm1"}
		vm2 := VM{UserID: "user", Name: "vm2"}
		vm3 := VM{UserID: "new-user", Name: "vm3"}

		err := db.CreateVM(&vm1)
		assert.NoError(t, err)
		err = db.CreateVM(&vm2)
		assert.NoError(t, err)
		err = db.CreateVM(&vm3)
		assert.NoError(t, err)

		vms, err := db.GetAllVms("user")
		assert.Equal(t, vms, []VM{vm1, vm2})
		assert.NoError(t, err)

		vms, err = db.GetAllVms("new-user")
		assert.Equal(t, vms, []VM{vm3})
		assert.NoError(t, err)

		err = db.DeleteAllVms("user")
		assert.NoError(t, err)

		vms, err = db.GetAllVms("user")
		assert.NoError(t, err)
		assert.Empty(t, vms)

		// other users unaffected
		vms, err = db.GetAllVms("new-user")
		assert.Equal(t, vms, []VM{vm3})
		assert.NoError(t, err)
	})
}

func TestCreateQuota(t *testing.T) {
	db := setupDB(t)
	quota := Quota{UserID: "user"}
	err := db.CreateQuota(&quota)
	assert.NoError(t, err)
	var q Quota
	err = db.db.First(&q).Error
	assert.NoError(t, err)
	assert.Equal(t, q, quota)
}

func TestUpdateUserQuota(t *testing.T) {
	db := setupDB(t)
	t.Run("quota not found so no updates", func(t *testing.T) {
		err := db.UpdateUserQuota("user", 5)
		assert.NoError(t, err)
	})
	t.Run("quota found", func(t *testing.T) {
		quota1 := Quota{UserID: "user"}
		quota2 := Quota{UserID: "new-user"}

		err := db.CreateQuota(&quota1)
		assert.NoError(t, err)
		err = db.CreateQuota(&quota2)
		assert.NoError(t, err)

		err = db.UpdateUserQuota("user", 5)
		assert.NoError(t, err)

		var q Quota
		err = db.db.First(&q, "user_id = 'user'").Error
		assert.NoError(t, err)
		assert.Equal(t, q.Vms, 5)

		err = db.db.First(&q, "user_id = 'new-user'").Error
		assert.NoError(t, err)
		assert.Equal(t, q.Vms, 0)

	})
}
func TestGetUserQuota(t *testing.T) {
	db := setupDB(t)
	t.Run("quota not found", func(t *testing.T) {
		_, err := db.GetUserQuota("user")
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("quota found", func(t *testing.T) {
		quota1 := Quota{UserID: "user"}
		quota2 := Quota{UserID: "new-user"}

		err := db.CreateQuota(&quota1)
		assert.NoError(t, err)
		err = db.CreateQuota(&quota2)
		assert.NoError(t, err)

		quota, err := db.GetUserQuota("user")
		assert.NoError(t, err)
		assert.Equal(t, quota, quota1)
	})
}

func TestCreateVoucher(t *testing.T) {
	db := setupDB(t)
	voucher := Voucher{UserID: "user"}
	err := db.CreateVoucher(&voucher)
	assert.NoError(t, err)
	var q Voucher
	err = db.db.First(&q).Error
	assert.NoError(t, err)
	assert.Equal(t, q, voucher)
}

func TestGetVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.GetVoucher("voucher")
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		voucher := Voucher{Voucher: "voucher"}
		err := db.CreateVoucher(&voucher)
		assert.NoError(t, err)

		v, err := db.GetVoucher("voucher")
		assert.Equal(t, v, voucher)
		assert.NoError(t, err)
	})
}
func TestGetVoucherByID(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.GetVoucherByID(1)
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		voucher := Voucher{Voucher: "voucher"}
		err := db.CreateVoucher(&voucher)
		assert.NoError(t, err)

		v, err := db.GetVoucherByID(voucher.ID)
		assert.Equal(t, v, voucher)
		assert.NoError(t, err)
	})
}

func TestListAllVouchers(t *testing.T) {
	db := setupDB(t)
	t.Run("vouchers not found", func(t *testing.T) {
		_, err := db.ListAllVouchers()
		assert.NoError(t, err)
	})
	t.Run("vouchers found", func(t *testing.T) {
		voucher1 := Voucher{Voucher: "voucher1", UserID: "user"}
		voucher2 := Voucher{Voucher: "voucher2", UserID: "new-user"}

		err := db.CreateVoucher(&voucher1)
		assert.NoError(t, err)
		err = db.CreateVoucher(&voucher2)
		assert.NoError(t, err)

		vouchers, err := db.ListAllVouchers()
		assert.NoError(t, err)
		assert.Equal(t, vouchers, []Voucher{voucher1, voucher2})
	})
}

func TestApproveVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.ApproveVoucher(1)
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		voucher1 := Voucher{Voucher: "voucher1", UserID: "user"}
		voucher2 := Voucher{Voucher: "voucher2", UserID: "new-user"}

		err := db.CreateVoucher(&voucher1)
		assert.NoError(t, err)
		err = db.CreateVoucher(&voucher2)
		assert.NoError(t, err)

		v, err := db.ApproveVoucher(voucher1.ID)
		assert.NoError(t, err)
		assert.Equal(t, v.Approved, true)

		var resVoucher Voucher
		err = db.db.First(&resVoucher, "user_id = 'user'").Error
		assert.Equal(t, v, resVoucher)
	})
}
func TestApproveAllVouchers(t *testing.T) {
	db := setupDB(t)
	t.Run("vouchers not found", func(t *testing.T) {
		_, err := db.ApproveAllVouchers()
		// missing where error because gorm uses the returned vouchers as the where clause
		// since no vouchers exist where clause is empty
		assert.Equal(t, err, gorm.ErrMissingWhereClause)
	})
	t.Run("vouchers found", func(t *testing.T) {
		voucher1 := Voucher{Voucher: "voucher1", UserID: "user"}
		voucher2 := Voucher{Voucher: "voucher2", UserID: "new-user"}

		err := db.CreateVoucher(&voucher1)
		assert.NoError(t, err)
		err = db.CreateVoucher(&voucher2)
		assert.NoError(t, err)

		v, err := db.ApproveAllVouchers()
		assert.NoError(t, err)
		assert.Len(t, v, 2)
		assert.True(t, v[0].Approved)
		assert.True(t, v[1].Approved)

		var vouchers []Voucher
		err = db.db.Find(&vouchers).Error
		assert.Equal(t, v, vouchers)
	})
}
func TestDeactivateVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found so no voucher updated", func(t *testing.T) {
		err := db.DeactivateVoucher("voucher")
		assert.NoError(t, err)
	})
	t.Run("vouchers found", func(t *testing.T) {
		voucher1 := Voucher{Voucher: "voucher1", UserID: "user"}
		voucher2 := Voucher{Voucher: "voucher2", UserID: "new-user"}

		err := db.CreateVoucher(&voucher1)
		assert.NoError(t, err)
		err = db.CreateVoucher(&voucher2)
		assert.NoError(t, err)

		err = db.DeactivateVoucher("voucher1")

		var v Voucher
		err = db.db.Find(&v).Where("voucher = 'voucher1'").Error
		assert.Equal(t, v.Used, true)
	})
}

func TestCreateK8s(t *testing.T) {
	db := setupDB(t)
	k8s := K8sCluster{
		UserID: "user",
		Master: Master{
			Name: "master",
		},
		Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
	}
	err := db.CreateK8s(&k8s)
	assert.NoError(t, err)
	var k K8sCluster
	err = db.db.First(&k).Error
	assert.NoError(t, err)
	assert.Equal(t, k.ID, 1)
	assert.Equal(t, k.UserID, "user")
	var m Master
	err = db.db.First(&m).Error
	assert.NoError(t, err)
	assert.Equal(t, m.Name, "master")
	assert.Equal(t, m.ClusterID, 1)
	var w []Worker
	err = db.db.Find(&w).Error
	assert.NoError(t, err)
	assert.Len(t, w, 2)
	assert.Equal(t, w[0].Name, "worker1")
	assert.Equal(t, w[0].ClusterID, 1)
	assert.Equal(t, w[1].Name, "worker2")
	assert.Equal(t, w[1].ClusterID, 1)
}
func TestGetK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		_, err := db.GetK8s(1)
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("K8s found", func(t *testing.T) {
		k8s := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}
		k8s2 := K8sCluster{
			UserID: "new-user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}

		err := db.CreateK8s(&k8s)
		assert.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		assert.NoError(t, err)

		k, err := db.GetK8s(k8s.ID)
		assert.NoError(t, err)
		assert.Equal(t, k, k8s)
		assert.NotEqual(t, k, k8s2)
	})
}
func TestGetAllK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		c, err := db.GetAllK8s("user")
		assert.NoError(t, err)
		assert.Empty(t, c)
	})
	t.Run("K8s found", func(t *testing.T) {
		k8s1 := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}
		k8s2 := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}
		k8s3 := K8sCluster{
			UserID: "new-user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}

		err := db.CreateK8s(&k8s1)
		assert.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		assert.NoError(t, err)
		err = db.CreateK8s(&k8s3)
		assert.NoError(t, err)

		k, err := db.GetAllK8s("user")
		assert.NoError(t, err)
		assert.Equal(t, k, []K8sCluster{k8s1, k8s2})

		k, err = db.GetAllK8s("new-user")
		assert.NoError(t, err)
		assert.Equal(t, k, []K8sCluster{k8s3})
	})
}
func TestDeleteK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		// unlike deleting vm it returns error because it find k8s from k8s table
		// and use it to filter master and workers
		err := db.DeleteK8s(1)
		assert.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("K8s found", func(t *testing.T) {
		k8s1 := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}
		k8s2 := K8sCluster{
			UserID: "new-user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}

		err := db.CreateK8s(&k8s1)
		assert.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		assert.NoError(t, err)

		err = db.DeleteK8s(k8s1.ID)
		assert.NoError(t, err)

		_, err = db.GetK8s(k8s1.ID)
		assert.Equal(t, err, gorm.ErrRecordNotFound)

		k, err := db.GetK8s(k8s2.ID)
		assert.NoError(t, err)
		assert.Equal(t, k, k8s2)
	})
}
func TestDeleteAllK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		// missing where error because gorm uses the returned clusters as the where clause
		// for deleting masters and workers since no clusters exist where clause is empty
		err := db.DeleteAllK8s("user")
		assert.Equal(t, err, gorm.ErrMissingWhereClause)
	})
	t.Run("K8s found", func(t *testing.T) {
		k8s1 := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}
		k8s2 := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}
		k8s3 := K8sCluster{
			UserID: "new-user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}

		err := db.CreateK8s(&k8s1)
		assert.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		assert.NoError(t, err)
		err = db.CreateK8s(&k8s3)
		assert.NoError(t, err)

		err = db.DeleteAllK8s("user")
		assert.NoError(t, err)

		k, err := db.GetAllK8s("user")
		assert.NoError(t, err)
		assert.Empty(t, k)

		k, err = db.GetAllK8s("new-user")
		assert.NoError(t, err)
		assert.Equal(t, k, []K8sCluster{k8s3})
	})
}
