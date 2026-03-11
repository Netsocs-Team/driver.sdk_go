package client

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-resty/resty/v2"
)

// ObjectGroup represents a group of objects/devices in Netsocs (DriverHub).
// It corresponds to the Group model documented in object_group.md.
type ObjectGroup struct {
	ID                   string                 `json:"id"`
	Name                 string                 `json:"name"`
	Type                 string                 `json:"type"` // "group" or "object"
	ItemID               string                 `json:"item_id"`
	Description          string                 `json:"description"`
	MainImage            string                 `json:"main_image"`
	GeoLocationPoints    []ObjectGroupGeoPoint  `json:"geo_location_points"`
	Multimedia           []ObjectGroupMedia     `json:"multimedia"`
	Color                string                 `json:"color"`
	Icon                 string                 `json:"icon"`
	AdditionalProperties map[string]interface{} `json:"additional_properties"`
	ParentID             string                 `json:"parent_id"`
	CreatedAt            string                 `json:"created_at"`
	UpdatedAt            string                 `json:"updated_at"`
}

// ObjectGroupGeoPoint is a geographic point associated with a group.
type ObjectGroupGeoPoint struct {
	Latitude  int `json:"latitude"`
	Longitude int `json:"longitude"`
}

// ObjectGroupMedia is a multimedia resource associated with a group.
type ObjectGroupMedia struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ObjectGroupTree is the recursive response from the full tree endpoint.
type ObjectGroupTree struct {
	Group    ObjectGroup       `json:"group"`
	Children []ObjectGroupTree `json:"children"`
}

// ObjectGroupTreeNode is the response from the direct children endpoint (lazy loading).
type ObjectGroupTreeNode struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	ItemID       string `json:"item_id"`
	Icon         string `json:"icon"`
	ChildrenLink string `json:"children_link"`
	GroupLink    string `json:"group_link"`
}

// CreateObjectGroupRequest is the body for creating an object group.
// Type must be "group" (contains other groups) or "object" (points to a device via ItemID).
type CreateObjectGroupRequest struct {
	Name                 string                 `json:"name"`
	Type                 string                 `json:"type"`
	ItemID               string                 `json:"item_id,omitempty"`
	Description          string                 `json:"description,omitempty"`
	MainImage            string                 `json:"main_image,omitempty"`
	GeoLocationPoints    []ObjectGroupGeoPoint  `json:"geo_location_points,omitempty"`
	Multimedia           []ObjectGroupMedia     `json:"multimedia,omitempty"`
	Color                string                 `json:"color,omitempty"`
	Icon                 string                 `json:"icon,omitempty"`
	AdditionalProperties map[string]interface{} `json:"additional_properties,omitempty"`
	ParentID             string                 `json:"parent_id,omitempty"`
}

// UpdateObjectGroupRequest is the body for updating a group (all fields are optional).
type UpdateObjectGroupRequest struct {
	Name                 string                 `json:"name,omitempty"`
	Type                 string                 `json:"type,omitempty"`
	ItemID               string                 `json:"item_id,omitempty"`
	Description          string                 `json:"description,omitempty"`
	MainImage            string                 `json:"main_image,omitempty"`
	GeoLocationPoints    []ObjectGroupGeoPoint  `json:"geo_location_points,omitempty"`
	Multimedia           []ObjectGroupMedia     `json:"multimedia,omitempty"`
	Color                string                 `json:"color,omitempty"`
	Icon                 string                 `json:"icon,omitempty"`
	AdditionalProperties map[string]interface{} `json:"additional_properties,omitempty"`
	ParentID             string                 `json:"parent_id,omitempty"`
}

// CreateObjectGroup creates a new object group in DriverHub.
// req.Type must be "group" or "object".
// parentID is optional: if provided, sets parent_id on the request.
func (c *NetsocsDriverClient) CreateObjectGroup(req CreateObjectGroupRequest, parentID ...string) (ObjectGroup, error) {
	if len(parentID) > 0 && parentID[0] != "" {
		req.ParentID = parentID[0]
	}
	url := c.driverHubHost + "/groups"
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", c.token).
		SetBody(req).
		Post(url)
	if err != nil {
		return ObjectGroup{}, err
	}
	if resp.IsError() {
		return ObjectGroup{}, fmt.Errorf("object-groups API create: %s", resp.String())
	}
	var group ObjectGroup
	if err := json.Unmarshal(resp.Body(), &group); err != nil {
		return ObjectGroup{}, err
	}
	return group, nil
}

// GetObjectGroup retrieves an object group by its ID.
func (c *NetsocsDriverClient) GetObjectGroup(id string) (ObjectGroup, error) {
	url := c.driverHubHost + "/groups/" + id
	resp, err := resty.New().R().
		SetHeader("X-Auth-Token", c.token).
		Get(url)
	if err != nil {
		return ObjectGroup{}, err
	}
	if resp.IsError() {
		return ObjectGroup{}, fmt.Errorf("object-groups API get: %s", resp.String())
	}
	var group ObjectGroup
	if err := json.Unmarshal(resp.Body(), &group); err != nil {
		return ObjectGroup{}, err
	}
	return group, nil
}

// UpdateObjectGroup updates an existing object group.
func (c *NetsocsDriverClient) UpdateObjectGroup(id string, req UpdateObjectGroupRequest) (ObjectGroup, error) {
	url := c.driverHubHost + "/groups/" + id
	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetHeader("X-Auth-Token", c.token).
		SetBody(req).
		Put(url)
	if err != nil {
		return ObjectGroup{}, err
	}
	if resp.IsError() {
		return ObjectGroup{}, errors.New(resp.String())
	}
	var group ObjectGroup
	if err := json.Unmarshal(resp.Body(), &group); err != nil {
		return ObjectGroup{}, err
	}
	return group, nil
}

// DeleteObjectGroup deletes (soft delete) an object group by ID.
func (c *NetsocsDriverClient) DeleteObjectGroup(id string) error {
	url := c.driverHubHost + "/groups/" + id
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

// GetObjectGroupTree returns the full group tree from the root,
// with all levels loaded recursively.
func (c *NetsocsDriverClient) GetObjectGroupTree() ([]ObjectGroupTree, error) {
	url := c.driverHubHost + "/groups/tree"
	resp, err := resty.New().R().
		SetHeader("X-Auth-Token", c.token).
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("object-groups API tree: %s", resp.String())
	}
	var result struct {
		Tree []ObjectGroupTree `json:"tree"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return result.Tree, nil
}

// GetObjectGroupChildren returns the direct children of a group (lazy loading).
// Each node includes navigation links (ChildrenLink, GroupLink).
func (c *NetsocsDriverClient) GetObjectGroupChildren(parentID string) ([]ObjectGroupTreeNode, error) {
	url := c.driverHubHost + "/groups/tree/children/" + parentID
	resp, err := resty.New().R().
		SetHeader("X-Auth-Token", c.token).
		Get(url)
	if err != nil {
		return nil, err
	}
	if resp.IsError() {
		return nil, fmt.Errorf("object-groups API children: %s", resp.String())
	}
	var result struct {
		Tree []ObjectGroupTreeNode `json:"tree"`
	}
	if err := json.Unmarshal(resp.Body(), &result); err != nil {
		return nil, err
	}
	return result.Tree, nil
}

// EnsureObjectGroup looks up a group of type "group" by name among the direct children
// of the given parentID (use "" to search at the root). If found, returns its ID;
// if it does not exist, creates it with the data from req and returns the new ID.
//
// Typical usage example in a driver:
//
//	groupID, err := client.EnsureObjectGroup("Building A", "", client.CreateObjectGroupRequest{
//	    Name: "Building A",
//	    Type: "group",
//	    Icon: "building",
//	})
func (c *NetsocsDriverClient) EnsureObjectGroup(name string, parentID string, req CreateObjectGroupRequest) (string, error) {
	var children []ObjectGroupTreeNode
	var err error

	if parentID == "" {
		// Search at root using the full tree to get only the first level
		tree, treeErr := c.GetObjectGroupTree()
		if treeErr != nil {
			return "", fmt.Errorf("EnsureObjectGroup: fetch tree: %w", treeErr)
		}
		for _, node := range tree {
			if node.Group.Name == name && node.Group.Type == "group" {
				return node.Group.ID, nil
			}
		}
	} else {
		children, err = c.GetObjectGroupChildren(parentID)
		if err != nil {
			return "", fmt.Errorf("EnsureObjectGroup: fetch children: %w", err)
		}
		for _, node := range children {
			if node.Name == name && node.Type == "group" {
				return node.ID, nil
			}
		}
	}

	// Not found: create the group
	req.Name = name
	req.Type = "group"
	if parentID != "" {
		req.ParentID = parentID
	}
	created, err := c.CreateObjectGroup(req)
	if err != nil {
		return "", fmt.Errorf("EnsureObjectGroup: create: %w", err)
	}
	return created.ID, nil
}
