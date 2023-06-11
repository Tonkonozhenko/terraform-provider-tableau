package tableau

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/resourcevalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"strings"
)

var (
	_ resource.Resource                = &projectPermissionResource{}
	_ resource.ResourceWithConfigure   = &projectPermissionResource{}
	_ resource.ResourceWithImportState = &projectPermissionResource{}
)

func NewProjectPermissionResource() resource.Resource {
	return &projectPermissionResource{}
}

type projectPermissionResource struct {
	client *Client
}

type projectPermissionResourceModel struct {
	ProjectID      types.String `tfsdk:"project_id"`
	GroupID        types.String `tfsdk:"group_id"`
	UserID         types.String `tfsdk:"user_id"`
	CapabilityName types.String `tfsdk:"capability_name"`
	CapabilityMode types.String `tfsdk:"capability_mode"`
}

func (r *projectPermissionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_permission"
}

func (r *projectPermissionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the project to remove the permission for.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the group to remove the permission for.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"user_id": schema.StringAttribute{
				Optional:    true,
				Description: "The ID of the user to remove the permission for.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"capability_name": schema.StringAttribute{
				Required:    true,
				Description: "The capability to remove the permission for.",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"Read",
						"Write",
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"capability_mode": schema.StringAttribute{
				Required:    true,
				Description: "Allow to remove the allow permission, or Deny to remove the deny permission. This value is case sensitive.",
				Validators: []validator.String{
					stringvalidator.OneOf([]string{
						"Allow",
						"Deny",
					}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *projectPermissionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectPermissionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectPermission := ProjectPermission{
		ProjectID:      plan.ProjectID.ValueString(),
		GroupID:        plan.GroupID.ValueString(),
		UserID:         plan.UserID.ValueString(),
		CapabilityName: plan.CapabilityName.ValueString(),
		CapabilityMode: plan.CapabilityMode.ValueString(),
	}

	err := r.client.AddProjectPermission(&projectPermission)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating project permissions",
			fmt.Sprintf("Could not create project permissions, unexpected error: %s", err.Error()),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectPermissionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectPermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	permissions, err := r.client.GetProjectPermissions(state.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Tableau Project Permissions",
			fmt.Sprintf("Could not read Tableau Project Permissions for ProjectID %s: %s", state.ProjectID.ValueString(), err.Error()),
		)
		return
	}

	userSearch := !state.UserID.IsNull()
	var found = false

	for _, permission := range *permissions {
		for _, capability := range permission.Capabilities.Capability {
			if capability.Name == state.CapabilityName.ValueString() && capability.Mode == state.CapabilityMode.ValueString() {
				if userSearch {
					if permission.User != nil && permission.User.ID == state.UserID.ValueString() {
						found = true
					}
				} else if permission.Group != nil && permission.Group.ID == state.GroupID.ValueString() {
					found = true
				}
			}
		}
	}

	if !found {
		resp.Diagnostics.AddError(
			"Error Finding Permission in existing permissions",
			"no permission in the existing permissions",
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectPermissionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	tflog.Info(ctx, "Project permissions do not support updates")
	return
}

func (r *projectPermissionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectPermissionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectPermission := ProjectPermission{
		ProjectID:      state.ProjectID.ValueString(),
		GroupID:        state.GroupID.ValueString(),
		UserID:         state.UserID.ValueString(),
		CapabilityName: state.CapabilityName.ValueString(),
		CapabilityMode: state.CapabilityMode.ValueString(),
	}
	err := r.client.DeleteProjectPermission(&projectPermission)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Tableau Project permissions",
			fmt.Sprintf("Could not project permissions, unexpected error: %s", err.Error()),
		)
		return
	}
}

func (r *projectPermissionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*Client)
}

func (r *projectPermissionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, ":")

	if len(idParts) != 5 {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: `project_id:group_id:user_id:capability_name:capability_mode`. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), idParts[0])...)
	if idParts[1] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_id"), idParts[1])...)
	}
	if idParts[2] != "" {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), idParts[2])...)
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("capability_name"), idParts[3])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("capability_mode"), idParts[4])...)
}

func (r *projectPermissionResource) ConfigValidators(ctx context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.Conflicting(
			path.MatchRoot("group_id"),
			path.MatchRoot("user_id"),
		),
	}
}
