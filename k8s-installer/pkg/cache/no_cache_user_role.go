package cache

import (
	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/log"
	"k8s-installer/schema"
)

type UserRoleNoCache struct {
	etcdConfig etcdClientConfig.EtcdConfig
}

func (noCachedUser *UserRoleNoCache) initialization() error {
	return nil
}

func (noCachedUser *UserRoleNoCache) CreateOrUpdateRole(roleId string, role schema.Role) error {
	return createOrUpdateRoleToDB(roleId, role, noCachedUser.etcdConfig)
}

func (noCachedUser *UserRoleNoCache) GetRoleList() (schema.RoleCollection, error) {
	return getRoleListFromDB(noCachedUser.etcdConfig)
}

func (noCachedUser *UserRoleNoCache) GetRole(roleId string) (*schema.Role, error) {
	roleList, err := getRoleListFromDB(noCachedUser.etcdConfig)
	if err != nil {
		return nil, err
	}
	for _, role := range roleList {
		if role.Id == roleId {
			return &schema.Role{
				Id:       role.Id,
				Name:     role.Name,
				Function: role.Function}, nil
		}
	}
	return nil, nil
}

func (noCachedUser *UserRoleNoCache) DeleteRole(roleId string) {
	err := deleteRoleFromDB(roleId, noCachedUser.etcdConfig)
	if err != nil {
		log.Error(err)
	}
}

func (noCachedUser *UserRoleNoCache) Login(username, password string) (*schema.User, error) {
	userList, err := getUserListFromDB(noCachedUser.etcdConfig)
	if err != nil {
		return nil, err
	}
	for _, user := range userList {
		if username == user.Username && password == user.Password {
			return &schema.User{
				Id:        user.Id,
				Username:  user.Username,
				Password:  user.Password,
				Roles:     user.Roles,
				Phone:     user.Phone,
				KsAccount: user.KsAccount,
			}, nil
		}
	}
	return nil, nil
}

func (noCachedUser *UserRoleNoCache) CreateOrUpdateUser(username string, user schema.User) error {
	return createOrUpdateUserToDB(username, user, noCachedUser.etcdConfig)
}

func (noCachedUser *UserRoleNoCache) GetUserList() (schema.UserCollection, error) {
	return getUserListFromDB(noCachedUser.etcdConfig)
}

func (noCachedUser *UserRoleNoCache) GetUser(username string) (*schema.User, error) {
	userList, err := getUserListFromDB(noCachedUser.etcdConfig)
	if err != nil {
		return nil, err
	}
	for _, user := range userList {
		if username == user.Username {
			return &schema.User{
				Id:        user.Id,
				Username:  user.Username,
				Password:  user.Password,
				Roles:     user.Roles,
				Phone:     user.Phone,
				KsAccount: user.KsAccount,
			}, nil
		}
	}
	return nil, nil
}

func (noCachedUser *UserRoleNoCache) DeleteUser(userId string) {
	err := deleteUserFromDB(userId, noCachedUser.etcdConfig)
	if err != nil {
		log.Error(err)
	}
}
