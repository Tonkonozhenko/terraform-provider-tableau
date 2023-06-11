package tableau

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &projectPermissionsDataSource{}
	_ datasource.DataSourceWithConfigure = &projectPermissionsDataSource{}
)

func ProjectPermissionsDataSource() datasource.DataSource {
	return &projectPermissionsDataSource{}
}

type projectPermissionsDataSource struct {
	client *Client
}

type projectPermissionsDataSourceModel struct {
	ProjectID           types.String         `tfsdk:"project_id"`
	GranteeCapabilities *[]GranteeCapability `tfsdk:"grantee_capabilities"`
}

func (d *projectPermissionsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_permissions"
}

func (d *projectPermissionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieve project permissions",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "ID of the project",
				Required:    true,
			},
			"grantee_capabilities": schema.ListNestedAttribute{
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"group": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required: true,
								},
							},
							Optional: true,
						},
						"user": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"id": schema.StringAttribute{
									Required: true,
								},
							},
							Optional: true,
						},
						"capabilities": schema.SingleNestedAttribute{
							Attributes: map[string]schema.Attribute{
								"capability": schema.ListNestedAttribute{
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Required: true,
											},
											"mode": schema.StringAttribute{
												Required: true,
											},
										},
									},
									Required: true,
								},
							},
							Required: true,
						},
					},
				},
				Computed: true,
			},
		},
	}
}

func (d *projectPermissionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectPermissionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	permissions, err := d.client.GetProjectPermissions(state.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Tableau Site",
			err.Error(),
		)
		return
	}

	state.GranteeCapabilities = permissions

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *projectPermissionsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	d.client = req.ProviderData.(*Client)
}
