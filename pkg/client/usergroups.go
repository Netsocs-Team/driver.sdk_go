package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// UserGroup representa un grupo de usuarios en Netsocs (DriverHub).
type UserGroup struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	UserIDs     []string `json:"user_ids"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// CreateUserGroupRequest es el cuerpo para crear un grupo de usuarios.
type CreateUserGroupRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	UserIDs     []string `json:"user_ids,omitempty"`
}

// UpdateUserGroupRequest es el cuerpo para actualizar un grupo (campos opcionales).
type UpdateUserGroupRequest struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	UserIDs     []string `json:"user_ids,omitempty"`
}

// CreateUserGroup crea un grupo de usuarios en el DriverHub.
// Ver usergroups.md para la documentación de la API.
func (c *NetsocsDriverClient) CreateUserGroup(req CreateUserGroupRequest) (UserGroup, error) {
	url := c.driverHubHost + "/user-groups"
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", c.token).
		SetBody(req).
		Post(url)
	if err != nil {
		return UserGroup{}, err
	}
	if resp.IsError() {
		return UserGroup{}, fmt.Errorf("user-groups API: %s", resp.String())
	}
	var group UserGroup
	if err := json.Unmarshal(resp.Body(), &group); err != nil {
		return UserGroup{}, err
	}
	return group, nil
}

// GetUserGroups lista todos los grupos de usuarios del sitio.
func (c *NetsocsDriverClient) GetUserGroups() ([]UserGroup, error) {
	url := c.driverHubHost + "/user-groups"
	resp, err := resty.New().R().
		SetHeader("X-Auth-Token", c.token).
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, errors.New(resp.String())
	}
	var groups []UserGroup
	if err := json.Unmarshal(resp.Body(), &groups); err != nil {
		return nil, err
	}
	return groups, nil
}

// GetUserGroup obtiene un grupo por ID.
func (c *NetsocsDriverClient) GetUserGroup(id string) (UserGroup, error) {
	url := c.driverHubHost + "/user-groups/" + id
	resp, err := resty.New().R().
		SetHeader("X-Auth-Token", c.token).
		Get(url)
	if err != nil {
		return UserGroup{}, err
	}
	if resp.IsError() {
		return UserGroup{}, errors.New(resp.String())
	}
	var group UserGroup
	if err := json.Unmarshal(resp.Body(), &group); err != nil {
		return UserGroup{}, err
	}
	return group, nil
}

// UpdateUserGroup actualiza un grupo de usuarios.
func (c *NetsocsDriverClient) UpdateUserGroup(id string, req UpdateUserGroupRequest) (UserGroup, error) {
	url := c.driverHubHost + "/user-groups/" + id
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", c.token).
		SetBody(req).
		Put(url)
	if err != nil {
		return UserGroup{}, err
	}
	if resp.IsError() {
		return UserGroup{}, errors.New(resp.String())
	}
	var group UserGroup
	if err := json.Unmarshal(resp.Body(), &group); err != nil {
		return UserGroup{}, err
	}
	return group, nil
}

// DeleteUserGroup elimina un grupo de usuarios.
func (c *NetsocsDriverClient) DeleteUserGroup(id string) error {
	url := c.driverHubHost + "/user-groups/" + id
	resp, err := resty.New().R().
		SetHeader("X-Auth-Token", c.token).
		Delete(url)
	if err != nil {
		return err
	}
	if resp.IsError() {
		return errors.New(resp.String())
	}
	return nil
}
