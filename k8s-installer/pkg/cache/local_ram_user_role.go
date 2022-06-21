package cache

import (
	"errors"
	"fmt"

	etcdClientConfig "k8s-installer/pkg/config/etcd_client"
	"k8s-installer/pkg/log"
	"k8s-installer/schema"
)

type UserRoleCache struct {
	users      schema.UserCollection
	etcdConfig etcdClientConfig.EtcdConfig
	roles      schema.RoleCollection
}

func (cachedUser *UserRoleCache) initialization() error {
	var err error
	cachedUser.users, err = getUserListFromDB(cachedUser.etcdConfig)
	if err != nil {
		return err
	}
	cachedUser.roles, err = getRoleListFromDB(cachedUser.etcdConfig)
	if err != nil {
		return err
	}
	return nil
}

func (cachedUser *UserRoleCache) Login(username, password string) (*schema.User, error) {
	userList, err := getUserListFromDB(cachedUser.etcdConfig)
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

func (cachedUser *UserRoleCache) CreateOrUpdateUser(username string, user schema.User) error {
	if err := createOrUpdateUserToDB(username, user, cachedUser.etcdConfig); err != nil {
		return err
	}
	cachedUser.users[user.Username] = user
	return nil
}

func (cachedUser *UserRoleCache) GetUserList() (schema.UserCollection, error) {
	return getUserListFromDB(cachedUser.etcdConfig)
}

func (cachedUser *UserRoleCache) GetUser(username string) (*schema.User, error) {
	if user, found := cachedUser.users[username]; !found {
		return nil, nil
	} else {
		return &user, nil
	}
}

func (cachedUser *UserRoleCache) DeleteUser(userId string) {
	err := deleteUserFromDB(userId, cachedUser.etcdConfig)
	if err != nil {
		log.Error(err)
	}
	delete(cachedUser.users, userId)
}

func (cachedUser *UserRoleCache) loadRole(user *schema.User) error {
	for index, role := range user.Roles {
		foundRole, err := cachedUser.GetRole(role.Id)
		if err != nil || foundRole == nil {
			errMsg := fmt.Sprintf("Failed to find role id %s of user %s", role.Id, user.Username)
			return errors.New(errMsg)
		}
		user.Roles[index] = *foundRole
	}
	return nil
}

func (cachedUser *UserRoleCache) CreateOrUpdateRole(roleId string, role schema.Role) error {
	if err := createOrUpdateRoleToDB(roleId, role, cachedUser.etcdConfig); err != nil {
		return err
	}
	cachedUser.roles[roleId] = role
	return nil
}

func (cachedUser *UserRoleCache) GetRoleList() (schema.RoleCollection, error) {
	return getRoleListFromDB(cachedUser.etcdConfig)
}

func (cachedUser *UserRoleCache) GetRole(roleId string) (*schema.Role, error) {
	if _, exists := cachedUser.roles[roleId]; exists {
		role := cachedUser.roles[roleId]
		return &role, nil
	}
	return nil, nil
}

func (cachedUser *UserRoleCache) DeleteRole(roleId string) {
	err := deleteRoleFromDB(roleId, cachedUser.etcdConfig)
	if err != nil {
		log.Error(err)
	}
	delete(cachedUser.roles, roleId)
}
