package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/abtme/terraform-provider-enom/internal/client"
)

const defaultBaseURL = "https://reseller.enom.com/interface.asp"

var _ provider.Provider = &EnomProvider{}

type EnomProvider struct {
	version string
}

type EnomProviderModel struct {
	BaseURL types.String `tfsdk:"base_url"`
	UID     types.String `tfsdk:"uid"`
	PW      types.String `tfsdk:"pw"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &EnomProvider{version: version}
	}
}

func (p *EnomProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "enom"
	resp.Version = p.version
}

func (p *EnomProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for managing eNom domain resources via the reseller API.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "eNom API base URL. Defaults to https://reseller.enom.com/interface.asp.",
			},
			"uid": schema.StringAttribute{
				Optional:    true,
				Description: "eNom reseller username. Can also be set via ENOM_UID environment variable.",
			},
			"pw": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "eNom reseller password. Can also be set via ENOM_PW environment variable.",
			},
		},
	}
}

func (p *EnomProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config EnomProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	baseURL := defaultBaseURL
	if !config.BaseURL.IsNull() && !config.BaseURL.IsUnknown() {
		baseURL = config.BaseURL.ValueString()
	}

	uid := os.Getenv("ENOM_UID")
	if !config.UID.IsNull() && !config.UID.IsUnknown() {
		uid = config.UID.ValueString()
	}
	if uid == "" {
		resp.Diagnostics.AddError("Missing uid", "uid must be set in provider config or ENOM_UID environment variable.")
		return
	}

	pw := os.Getenv("ENOM_PW")
	if !config.PW.IsNull() && !config.PW.IsUnknown() {
		pw = config.PW.ValueString()
	}
	if pw == "" {
		resp.Diagnostics.AddError("Missing pw", "pw must be set in provider config or ENOM_PW environment variable.")
		return
	}

	c := client.NewClient(baseURL, uid, pw)
	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *EnomProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDomainNameserversResource,
	}
}

func (p *EnomProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}
