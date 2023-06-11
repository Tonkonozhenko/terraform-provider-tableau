package tableau

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Capability struct {
	Name string `json:"name" tfsdk:"name"`
	Mode string `json:"mode" tfsdk:"mode"`
}

type CapabilitiesWrapper struct {
	Capability []Capability `json:"capability" tfsdk:"capability"`
}

type GranteeCapability struct {
	Group        *Owner              `json:"group,omitempty" tfsdk:"group"`
	User         *Owner              `json:"user,omitempty" tfsdk:"user"`
	Capabilities CapabilitiesWrapper `json:"capabilities" tfsdk:"capabilities"`
}

type ProjectPermissionsResponse struct {
	Project             *Project            `json:"project,omitempty"`
	GranteeCapabilities []GranteeCapability `json:"granteeCapabilities"`
}

type ProjectPermissionsWrapperResponse struct {
	Permissions ProjectPermissionsResponse `json:"permissions"`
}

type ProjectPermission struct {
	ProjectID      string `tfsdk:"project_id"`
	GroupID        string `tfsdk:"group_id"`
	UserID         string `tfsdk:"user_id"`
	CapabilityName string `tfsdk:"capability_name"`
	CapabilityMode string `tfsdk:"capability_mode"`
}

func (c *Client) GetProjectPermissions(projectID string) (*[]GranteeCapability, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/projects/%s/permissions", c.ApiUrl, projectID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	userResponse := ProjectPermissionsWrapperResponse{}
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		return nil, err
	}

	return &userResponse.Permissions.GranteeCapabilities, nil
}

func (c *Client) AddProjectPermission(p *ProjectPermission) error {
	capabilities := CapabilitiesWrapper{
		Capability: []Capability{
			{
				Name: p.CapabilityName,
				Mode: p.CapabilityMode,
			},
		},
	}

	var gc GranteeCapability
	if p.GroupID != "" {
		gc = GranteeCapability{
			Group:        &Owner{ID: p.GroupID},
			Capabilities: capabilities,
		}
	} else {
		gc = GranteeCapability{
			User:         &Owner{ID: p.UserID},
			Capabilities: capabilities,
		}
	}

	reqBody := ProjectPermissionsWrapperResponse{
		Permissions: ProjectPermissionsResponse{
			GranteeCapabilities: []GranteeCapability{gc},
		},
	}

	reqJson, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/projects/%s/permissions", c.ApiUrl, p.ProjectID)
	req, err := http.NewRequest("PUT", url, strings.NewReader(string(reqJson)))
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteProjectPermission(p *ProjectPermission) error {
	var url string
	if p.GroupID != "" {
		url = fmt.Sprintf("%s/projects/%s/permissions/groups/%s/%s/%s", c.ApiUrl, p.ProjectID, p.GroupID, p.CapabilityName, p.CapabilityMode)
	} else {
		url = fmt.Sprintf("%s/projects/%s/permissions/users/%s/%s/%s", c.ApiUrl, p.ProjectID, p.UserID, p.CapabilityName, p.CapabilityMode)
	}

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	_, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}
