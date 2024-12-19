// Package models for database models
package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupDB(t *testing.T) DB {
	db := NewDB()
	testDir := t.TempDir()

	dbName := "test.db"
	err := db.Connect(testDir + dbName)
	require.NoError(t, err)
	err = db.Migrate()
	require.NoError(t, err)
	return db
}

func TestConnect(t *testing.T) {
	db := NewDB()
	testDir := t.TempDir()
	dbName := "test.db"
	t.Run("invalid path", func(t *testing.T) {
		err := db.Connect(testDir + "/another_dir/" + dbName)
		require.Error(t, err)
	})
	t.Run("valid path", func(t *testing.T) {
		err := db.Connect(testDir + dbName)
		require.NoError(t, err)
	})
}

func TestCreateUser(t *testing.T) {
	db := setupDB(t)
	err := db.CreateUser(&User{
		FirstName: "test",
	})
	require.NoError(t, err)
	var user User
	err = db.db.First(&user).Error
	require.Equal(t, user.FirstName, "test")
	require.NoError(t, err)
}

func TestGetUserByEmail(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found", func(t *testing.T) {
		err := db.CreateUser(&User{
			FirstName: "test",
		})
		require.NoError(t, err)
		_, err = db.GetUserByEmail("email")
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("user found", func(t *testing.T) {
		err := db.CreateUser(&User{
			FirstName: "test",
			Email:     "email",
		})
		require.NoError(t, err)
		u, err := db.GetUserByEmail("email")
		require.Equal(t, u.FirstName, "test")
		require.Equal(t, u.Email, "email")
		require.NoError(t, err)
	})
}

func TestGetUserByID(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found", func(t *testing.T) {
		err := db.CreateUser(&User{
			FirstName: "test",
		})
		require.NoError(t, err)
		_, err = db.GetUserByID("not-uuid")
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			FirstName: "test",
			Email:     "email",
		}
		err := db.CreateUser(&user)
		require.NoError(t, err)
		u, err := db.GetUserByID(user.ID.String())
		require.Equal(t, u.FirstName, "test")
		require.Equal(t, u.Email, "email")
		require.NoError(t, err)
	})
}

func TestListAllUsers(t *testing.T) {
	db := setupDB(t)
	t.Run("no users in list", func(t *testing.T) {
		users, err := db.ListAllUsers()
		require.NoError(t, err)
		require.Empty(t, users)
	})

	t.Run("list all users for admin", func(t *testing.T) {
		user1 := User{
			FirstName:      "user1",
			Email:          "user1@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
		}

		err := db.CreateUser(&user1)
		require.NoError(t, err)
		users, err := db.ListAllUsers()
		require.NoError(t, err)
		require.Equal(t, users[0].FirstName, user1.FirstName)
		require.Equal(t, users[0].Email, user1.Email)
		require.Equal(t, users[0].HashedPassword, user1.HashedPassword)

	})
}

func TestGetCodeByEmail(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found", func(t *testing.T) {
		_, err := db.GetCodeByEmail("email@gmail.com")
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})

	t.Run("get code of user", func(t *testing.T) {
		user := User{
			FirstName:      "user",
			Email:          "user@gmail.com",
			HashedPassword: []byte{},
			Verified:       true,
			Code:           1234,
		}

		err := db.CreateUser(&user)
		require.NoError(t, err)
		code, err := db.GetCodeByEmail("user@gmail.com")
		require.NoError(t, err)
		require.Equal(t, code, user.Code)
	})

}

func TestUpdatePassword(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found so nothing updated", func(t *testing.T) {
		err := db.UpdateUserPassword("email", []byte("new-pass"))
		require.Error(t, err)
		var user User
		err = db.db.First(&user).Error
		require.Equal(t, err, gorm.ErrRecordNotFound)
		require.Empty(t, user)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email:          "email",
			HashedPassword: []byte("new-pass"),
		}
		err := db.CreateUser(&user)
		require.NoError(t, err)
		err = db.UpdateUserPassword("email", []byte("new-pass"))
		require.NoError(t, err)
		u, err := db.GetUserByEmail("email")
		require.Equal(t, u.Email, "email")
		require.Equal(t, u.HashedPassword, []byte("new-pass"))
		require.NoError(t, err)
	})
}

func TestUpdateUserByID(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found so nothing updated", func(t *testing.T) {
		err := db.UpdateUserByID(User{Email: "email"})
		require.NoError(t, err)
		var user User
		err = db.db.First(&user).Error
		require.Equal(t, err, gorm.ErrRecordNotFound)
		require.Empty(t, user)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email:          "email",
			HashedPassword: []byte{},
		}
		err := db.CreateUser(&user)
		require.NoError(t, err)
		err = db.UpdateUserByID(User{
			ID:             user.ID,
			Email:          "",
			HashedPassword: []byte("new-pass"),
			FirstName:      "name",
		})
		require.NoError(t, err)
		var u User
		err = db.db.First(&u).Error
		// shouldn't change
		require.Equal(t, u.Email, user.Email)
		// should change
		require.Equal(t, u.HashedPassword, []byte("new-pass"))
		require.Equal(t, u.FirstName, "name")

		require.NoError(t, err)
	})
}

func TestUpdateVerification(t *testing.T) {
	db := setupDB(t)
	t.Run("user not found so nothing updated", func(t *testing.T) {
		err := db.UpdateUserVerification("id", true)
		require.NoError(t, err)
		var user User
		err = db.db.First(&user).Error
		require.Equal(t, err, gorm.ErrRecordNotFound)
		require.Empty(t, user)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email: "email",
		}
		err := db.CreateUser(&user)
		require.Equal(t, user.Verified, false)
		require.NoError(t, err)
		err = db.UpdateUserVerification(user.ID.String(), true)
		require.NoError(t, err)
		var u User
		err = db.db.First(&u).Error
		require.NoError(t, err)
		require.Equal(t, u.Verified, true)
	})
}
func TestAddUserVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("user and voucher not found so nothing updated", func(t *testing.T) {
		err := db.DeactivateVoucher("id", "voucher")
		require.NoError(t, err)
		var user User
		var voucher Voucher

		err = db.db.First(&user).Error
		require.Equal(t, err, gorm.ErrRecordNotFound)
		require.Empty(t, user)

		err = db.db.First(&voucher).Error
		require.Equal(t, err, gorm.ErrRecordNotFound)
		require.Empty(t, voucher)
	})
	t.Run("user found", func(t *testing.T) {
		user := User{
			Email: "email",
		}
		voucher := Voucher{
			Voucher: "voucher",
		}
		err := db.CreateUser(&user)
		require.NoError(t, err)
		err = db.db.Create(&voucher).Error
		require.NoError(t, err)
		require.Equal(t, voucher.Used, false)

		err = db.DeactivateVoucher(user.ID.String(), "voucher")
		require.NoError(t, err)
		var u User
		var v Voucher
		err = db.db.First(&u).Error
		require.NoError(t, err)

		err = db.db.First(&v).Error
		require.NoError(t, err)
		require.Equal(t, v.Used, true)
	})
}

func TestGetNotUsedVoucherByUserID(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.GetNotUsedVoucherByUserID("id")
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher is used", func(t *testing.T) {
		user := User{
			Email: "email1",
		}
		err := db.CreateUser(&user)
		require.NoError(t, err)
		voucher := Voucher{
			UserID: user.ID.String(),
			Used:   true,
		}

		err = db.db.Create(&voucher).Error
		require.NoError(t, err)

		_, err = db.GetNotUsedVoucherByUserID(user.ID.String())
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		user := User{
			Email: "email2",
		}
		err := db.CreateUser(&user)
		require.NoError(t, err)
		voucher := Voucher{
			UserID:  user.ID.String(),
			Voucher: "voucher2",
			Used:    false,
		}

		err = db.db.Create(&voucher).Error
		require.NoError(t, err)

		v, err := db.GetNotUsedVoucherByUserID(user.ID.String())
		require.NoError(t, err)
		voucher.CreatedAt = v.CreatedAt
		voucher.UpdatedAt = v.UpdatedAt
		require.Equal(t, voucher, v)
	})
}

func TestCreateVM(t *testing.T) {
	db := setupDB(t)
	vm := VM{Name: "vm"}
	err := db.CreateVM(&vm)
	require.NoError(t, err)

	var v VM
	require.NoError(t, db.db.First(&v).Error)

	v.CreatedAt = time.Now().In(time.Local).Truncate(time.Second)
	vm.CreatedAt = time.Now().Truncate(time.Second)
	require.Equal(t, v, vm)
}

func TestGetVMByID(t *testing.T) {
	db := setupDB(t)
	t.Run("vm not found", func(t *testing.T) {
		_, err := db.GetVMByID(1)
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("vm found", func(t *testing.T) {
		vm := VM{Name: "vm"}
		err := db.CreateVM(&vm)
		require.NoError(t, err)

		v, err := db.GetVMByID(vm.ID)
		require.NoError(t, err)

		v.CreatedAt = time.Now().In(time.Local).Truncate(time.Second)
		vm.CreatedAt = time.Now().Truncate(time.Second)
		require.Equal(t, v, vm)
	})
}

func TestGetAllVMs(t *testing.T) {
	db := setupDB(t)
	t.Run("no vms with user", func(t *testing.T) {
		vms, err := db.GetAllVms("user")
		require.NoError(t, err)
		require.Empty(t, vms)
	})
	t.Run("vms for different users", func(t *testing.T) {
		vm1 := VM{UserID: "user", Name: "vm1"}
		vm2 := VM{UserID: "user", Name: "vm2"}
		vm3 := VM{UserID: "new-user", Name: "vm3"}

		err := db.CreateVM(&vm1)
		require.NoError(t, err)
		err = db.CreateVM(&vm2)
		require.NoError(t, err)
		err = db.CreateVM(&vm3)
		require.NoError(t, err)

		vms, err := db.GetAllVms("user")

		vms[0].CreatedAt = time.Now().In(time.Local).Truncate(time.Second)
		vms[1].CreatedAt = time.Now().In(time.Local).Truncate(time.Second)
		vm1.CreatedAt = time.Now().Truncate(time.Second)
		vm2.CreatedAt = time.Now().Truncate(time.Second)

		require.Equal(t, vms, []VM{vm1, vm2})
		require.NoError(t, err)

		vms, err = db.GetAllVms("new-user")
		vms[0].CreatedAt = time.Now().In(time.Local).Truncate(time.Second)
		vm3.CreatedAt = time.Now().Truncate(time.Second)

		require.Equal(t, vms, []VM{vm3})
		require.NoError(t, err)
	})
}

func TestAvailableVMName(t *testing.T) {
	db := setupDB(t)
	t.Run("no vms", func(t *testing.T) {
		valid, err := db.AvailableVMName("user")
		require.NoError(t, err)
		require.Empty(t, false, valid)
	})

	t.Run("test with existing name", func(t *testing.T) {
		vm := VM{UserID: "user", Name: "vm1"}
		err := db.CreateVM(&vm)
		require.NoError(t, err)

		valid, err := db.AvailableVMName("vm1")
		require.NoError(t, err)
		require.Equal(t, false, valid)

	})

	t.Run("test with new name", func(t *testing.T) {
		vm := VM{UserID: "user", Name: "vm2"}
		err := db.CreateVM(&vm)
		require.NoError(t, err)

		valid, err := db.AvailableVMName("vm")
		require.NoError(t, err)
		require.Equal(t, true, valid)

	})
}

func TestDeleteVMByID(t *testing.T) {
	db := setupDB(t)
	t.Run("delete non existing vm", func(t *testing.T) {
		// gorm doesn't return error if vm doesn't exist
		err := db.DeleteVMByID(1)
		require.NoError(t, err)
	})
	t.Run("delete existing vm", func(t *testing.T) {
		vm := VM{UserID: "user", Name: "vm"}
		err := db.CreateVM(&vm)
		require.NoError(t, err)

		err = db.DeleteVMByID(vm.ID)
		require.NoError(t, err)

		var v VM
		err = db.db.First(&v).Error
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
}

func TestDeleteAllVMs(t *testing.T) {
	db := setupDB(t)
	t.Run("delete non existing vms", func(t *testing.T) {
		// gorm doesn't return error if vms don't exist
		err := db.DeleteAllVms("user")
		require.NoError(t, err)
	})

	t.Run("delete existing vms", func(t *testing.T) {
		vm1 := VM{UserID: "user", Name: "vm1"}
		vm2 := VM{UserID: "user", Name: "vm2"}
		vm3 := VM{UserID: "new-user", Name: "vm3"}

		err := db.CreateVM(&vm1)
		require.NoError(t, err)
		err = db.CreateVM(&vm2)
		require.NoError(t, err)
		err = db.CreateVM(&vm3)
		require.NoError(t, err)

		err = db.DeleteAllVms("user")
		require.NoError(t, err)

		vms, err := db.GetAllVms("user")
		require.NoError(t, err)
		require.Empty(t, vms)

		// other users unaffected
		vms, err = db.GetAllVms("new-user")
		vms[0].CreatedAt = time.Now().In(time.Local).Truncate(time.Second)
		vm3.CreatedAt = time.Now().Truncate(time.Second)
		require.Equal(t, vms, []VM{vm3})
		require.NoError(t, err)
	})
}

func TestCreateVoucher(t *testing.T) {
	db := setupDB(t)
	voucher := Voucher{UserID: "user"}
	err := db.CreateVoucher(&voucher)
	require.NoError(t, err)
	var q Voucher
	err = db.db.First(&q).Error
	require.NoError(t, err)
	voucher.CreatedAt = q.CreatedAt
	voucher.UpdatedAt = q.UpdatedAt
	require.Equal(t, q, voucher)
}

func TestGetVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.GetVoucher("voucher")
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		voucher := Voucher{Voucher: "voucher"}
		err := db.CreateVoucher(&voucher)
		require.NoError(t, err)

		v, err := db.GetVoucher("voucher")
		voucher.CreatedAt = v.CreatedAt
		voucher.UpdatedAt = v.UpdatedAt
		require.Equal(t, v, voucher)
		require.NoError(t, err)
	})
}
func TestGetVoucherByID(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.GetVoucherByID(1)
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		voucher := Voucher{Voucher: "voucher"}
		err := db.CreateVoucher(&voucher)
		require.NoError(t, err)

		v, err := db.GetVoucherByID(voucher.ID)
		voucher.CreatedAt = v.CreatedAt
		voucher.UpdatedAt = v.UpdatedAt
		require.Equal(t, v, voucher)
		require.NoError(t, err)
	})
}

func TestListAllVouchers(t *testing.T) {
	db := setupDB(t)
	t.Run("vouchers not found", func(t *testing.T) {
		_, err := db.ListAllVouchers()
		require.NoError(t, err)
	})
	t.Run("vouchers found", func(t *testing.T) {
		voucher1 := Voucher{Voucher: "voucher1", UserID: "user"}
		voucher2 := Voucher{Voucher: "voucher2", UserID: "new-user"}

		err := db.CreateVoucher(&voucher1)
		require.NoError(t, err)
		err = db.CreateVoucher(&voucher2)
		require.NoError(t, err)

		vouchers, err := db.ListAllVouchers()
		require.NoError(t, err)
		require.Equal(t, len(vouchers), 2)
	})
}

func TestApproveVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found", func(t *testing.T) {
		_, err := db.UpdateVoucher(1, true)
		require.Equal(t, err, gorm.ErrRecordNotFound)
	})
	t.Run("voucher found", func(t *testing.T) {
		voucher1 := Voucher{Voucher: "voucher1", UserID: "user"}
		voucher2 := Voucher{Voucher: "voucher2", UserID: "new-user"}

		err := db.CreateVoucher(&voucher1)
		require.NoError(t, err)
		err = db.CreateVoucher(&voucher2)
		require.NoError(t, err)

		v, err := db.UpdateVoucher(voucher1.ID, true)
		require.NoError(t, err)
		require.True(t, v.Approved)

		var resVoucher Voucher
		err = db.db.First(&resVoucher, "user_id = 'user'").Error
		require.NoError(t, err)
		resVoucher.CreatedAt = v.CreatedAt
		resVoucher.UpdatedAt = v.UpdatedAt
		require.Equal(t, v, resVoucher)
	})
}

func TestDeactivateVoucher(t *testing.T) {
	db := setupDB(t)
	t.Run("voucher not found so no voucher updated", func(t *testing.T) {
		err := db.DeactivateVoucher("user", "voucher")
		require.NoError(t, err)
	})
	t.Run("vouchers found", func(t *testing.T) {
		voucher1 := Voucher{Voucher: "voucher1", UserID: "user"}
		voucher2 := Voucher{Voucher: "voucher2", UserID: "new-user"}

		err := db.CreateVoucher(&voucher1)
		require.NoError(t, err)
		err = db.CreateVoucher(&voucher2)
		require.NoError(t, err)

		err = db.DeactivateVoucher("user", "voucher1")
		require.NoError(t, err)

		var v Voucher
		err = db.db.Find(&v).Where("voucher = 'voucher1'").Error
		require.NoError(t, err)
		require.Equal(t, v.Used, true)
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
	require.NoError(t, err)
	var k K8sCluster
	err = db.db.First(&k).Error
	require.NoError(t, err)
	require.Equal(t, k.ID, 1)
	require.Equal(t, k.UserID, "user")
	var m Master
	err = db.db.First(&m).Error
	require.NoError(t, err)
	require.Equal(t, m.Name, "master")
	require.Equal(t, m.ClusterID, 1)
	var w []Worker
	err = db.db.Find(&w).Error
	require.NoError(t, err)
	require.Len(t, w, 2)
	require.Equal(t, w[0].Name, "worker1")
	require.Equal(t, w[0].ClusterID, 1)
	require.Equal(t, w[1].Name, "worker2")
	require.Equal(t, w[1].ClusterID, 1)
}

func TestGetK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		_, err := db.GetK8s(1)
		require.Equal(t, err, gorm.ErrRecordNotFound)
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
				Name: "master2",
			},
			Workers: []Worker{{Name: "worker3"}, {Name: "worker4"}},
		}

		err := db.CreateK8s(&k8s)
		require.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		require.NoError(t, err)

		k, err := db.GetK8s(k8s.ID)
		require.NoError(t, err)
		require.Equal(t, len(k.Workers), len(k8s.Workers))
		require.Equal(t, k.Master.ClusterID, k8s.Master.ClusterID)
		require.NotEqual(t, k, k8s2)
	})
}

func TestGetAllK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		c, err := db.GetAllK8s("user")
		require.NoError(t, err)
		require.Empty(t, c)
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
				Name: "master2",
			},
			Workers: []Worker{{Name: "worker3"}, {Name: "worker4"}},
		}
		k8s3 := K8sCluster{
			UserID: "new-user",
			Master: Master{
				Name: "master3",
			},
			Workers: []Worker{{Name: "worker5"}, {Name: "worker6"}},
		}

		err := db.CreateK8s(&k8s1)
		require.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		require.NoError(t, err)
		err = db.CreateK8s(&k8s3)
		require.NoError(t, err)

		k, err := db.GetAllK8s("user")
		require.NoError(t, err)
		require.Equal(t, len(k[0].Workers), len(k8s1.Workers))
		require.Equal(t, k[0].Master.ClusterID, k8s1.Master.ClusterID)
		require.Equal(t, len(k[1].Workers), len(k8s2.Workers))
		require.Equal(t, k[1].Master.ClusterID, k8s2.Master.ClusterID)

		k, err = db.GetAllK8s("new-user")
		require.NoError(t, err)
		require.Equal(t, len(k[0].Workers), len(k8s3.Workers))
		require.Equal(t, k[0].Master.ClusterID, k8s3.Master.ClusterID)
	})

}

func TestDeleteK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		// unlike deleting vm it returns error because it find k8s from k8s table
		// and use it to filter master and workers
		err := db.DeleteK8s(1)
		require.Equal(t, err, gorm.ErrRecordNotFound)
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
				Name: "master2",
			},
			Workers: []Worker{{Name: "worker3"}, {Name: "worker4"}},
		}

		err := db.CreateK8s(&k8s1)
		require.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		require.NoError(t, err)

		err = db.DeleteK8s(k8s1.ID)
		require.NoError(t, err)

		_, err = db.GetK8s(k8s1.ID)
		require.Equal(t, err, gorm.ErrRecordNotFound)

		k, err := db.GetK8s(k8s2.ID)
		require.NoError(t, err)
		require.Equal(t, len(k.Workers), len(k8s2.Workers))
		require.Equal(t, k.Master.ClusterID, k8s2.Master.ClusterID)
	})
}

func TestDeleteAllK8s(t *testing.T) {
	db := setupDB(t)
	t.Run("K8s not found", func(t *testing.T) {
		// missing where error because gorm uses the returned clusters as the where clause
		// for deleting masters and workers since no clusters exist where clause is empty
		err := db.DeleteAllK8s("user")
		require.Equal(t, err, gorm.ErrMissingWhereClause)
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
				Name: "master2",
			},
			Workers: []Worker{{Name: "worker3"}, {Name: "worker4"}},
		}
		k8s3 := K8sCluster{
			UserID: "new-user",
			Master: Master{
				Name: "master3",
			},
			Workers: []Worker{{Name: "worker5"}, {Name: "worker6"}},
		}

		err := db.CreateK8s(&k8s1)
		require.NoError(t, err)
		err = db.CreateK8s(&k8s2)
		require.NoError(t, err)
		err = db.CreateK8s(&k8s3)
		require.NoError(t, err)

		err = db.DeleteAllK8s("user")
		require.NoError(t, err)

		k, err := db.GetAllK8s("user")
		require.NoError(t, err)
		require.Empty(t, k)

		k, err = db.GetAllK8s("new-user")
		require.NoError(t, err)

		require.Equal(t, len(k[0].Workers), len(k8s3.Workers))
		require.Equal(t, k[0].Master.ClusterID, k8s3.Master.ClusterID)
	})

	t.Run("test with no id", func(t *testing.T) {
		err := db.DeleteAllK8s("")
		require.Error(t, err)
	})
}

func TestAvailableK8sName(t *testing.T) {
	db := setupDB(t)
	t.Run("no k8s", func(t *testing.T) {
		valid, err := db.AvailableK8sName("k8s")
		require.NoError(t, err)
		require.Empty(t, false, valid)
	})

	t.Run("test with existing name", func(t *testing.T) {
		k8s := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master",
			},
			Workers: []Worker{{Name: "worker1"}, {Name: "worker2"}},
		}
		err := db.CreateK8s(&k8s)
		require.NoError(t, err)

		valid, err := db.AvailableK8sName("master")
		require.NoError(t, err)
		require.Equal(t, false, valid)

	})

	t.Run("test with new name", func(t *testing.T) {
		k8s := K8sCluster{
			UserID: "user",
			Master: Master{
				Name: "master2",
			},
			Workers: []Worker{{Name: "worker3"}, {Name: "worker4"}},
		}
		err := db.CreateK8s(&k8s)
		require.NoError(t, err)

		valid, err := db.AvailableK8sName("new-master")
		require.NoError(t, err)
		require.Equal(t, true, valid)

	})

}

func TestUpdateMaintenance(t *testing.T) {
	db := setupDB(t)
	err := db.UpdateMaintenance(true)
	require.NoError(t, err)

}

func TestGetMaintenance(t *testing.T) {
	db := setupDB(t)
	err := db.UpdateMaintenance(true)
	require.NoError(t, err)

	m, err := db.GetMaintenance()
	require.NoError(t, err)
	require.True(t, m.Active)
}

func TestUpdateNextLaunch(t *testing.T) {
	db := setupDB(t)
	err := db.UpdateNextLaunch(true)
	require.NoError(t, err)

}

func TestGetNextLaunch(t *testing.T) {
	db := setupDB(t)
	err := db.UpdateNextLaunch(true)
	require.NoError(t, err)

	m, err := db.GetNextLaunch()
	require.NoError(t, err)
	require.True(t, m.Launched)
}
