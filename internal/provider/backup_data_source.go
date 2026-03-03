package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &BackupDataSource{}

type BackupDataSource struct {
	client *AutoLoginClient
}

type BackupDataSourceModel struct {
	Backups types.List `tfsdk:"backups"`
}

var backupAttrTypes = map[string]attr.Type{
	"controller_name": types.StringType,
	"filename":        types.StringType,
	"type":            types.StringType,
	"version":         types.StringType,
	"time":            types.Int64Type,
	"datetime":        types.StringType,
	"format":          types.StringType,
	"days":            types.Int64Type,
	"keep_forever":    types.BoolType,
	"note":            types.StringType,
	"size":            types.Int64Type,
}

func NewBackupDataSource() datasource.DataSource {
	return &BackupDataSource{}
}

func (d *BackupDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (d *BackupDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists all available backups on the UniFi controller.",
		Attributes: map[string]schema.Attribute{
			"backups": schema.ListNestedAttribute{
				Description: "List of backups.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"controller_name": schema.StringAttribute{Computed: true},
						"filename":        schema.StringAttribute{Computed: true},
						"type":            schema.StringAttribute{Computed: true},
						"version":         schema.StringAttribute{Computed: true},
						"time":            schema.Int64Attribute{Computed: true},
						"datetime":        schema.StringAttribute{Computed: true},
						"format":          schema.StringAttribute{Computed: true},
						"days":            schema.Int64Attribute{Computed: true},
						"keep_forever":    schema.BoolAttribute{Computed: true},
						"note":            schema.StringAttribute{Computed: true},
						"size":            schema.Int64Attribute{Computed: true},
					},
				},
			},
		},
	}
}

func (d *BackupDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*AutoLoginClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *AutoLoginClient, got: %T.", req.ProviderData),
		)
		return
	}
	d.client = client
}

func (d *BackupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	backups, err := d.client.ListBackups(ctx)
	if err != nil {
		handleSDKError(&resp.Diagnostics, err, "list", "backups")
		return
	}

	var state BackupDataSourceModel
	backupValues := make([]attr.Value, len(backups))
	for i, b := range backups {
		vals := map[string]attr.Value{
			"controller_name": stringValueOrNull(b.ControllerName),
			"filename":        stringValueOrNull(b.Filename),
			"type":            stringValueOrNull(b.Type),
			"version":         stringValueOrNull(b.Version),
			"datetime":        stringValueOrNull(b.Datetime),
			"format":          stringValueOrNull(b.Format),
			"note":            stringValueOrNull(b.Note),
			"keep_forever":    types.BoolValue(derefBool(b.KeepForever)),
		}
		if b.Time != nil {
			vals["time"] = types.Int64Value(*b.Time)
		} else {
			vals["time"] = types.Int64Null()
		}
		if b.Days != nil {
			vals["days"] = types.Int64Value(int64(*b.Days))
		} else {
			vals["days"] = types.Int64Null()
		}
		if b.Size != nil {
			vals["size"] = types.Int64Value(*b.Size)
		} else {
			vals["size"] = types.Int64Null()
		}

		obj, diags := types.ObjectValue(backupAttrTypes, vals)
		resp.Diagnostics.Append(diags...)
		backupValues[i] = obj
	}

	if resp.Diagnostics.HasError() {
		return
	}

	list, diags := types.ListValue(types.ObjectType{AttrTypes: backupAttrTypes}, backupValues)
	resp.Diagnostics.Append(diags...)
	state.Backups = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
