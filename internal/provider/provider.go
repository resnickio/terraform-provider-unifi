package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/resnickio/unifi-go-sdk/pkg/unifi"
)

var _ provider.Provider = &UnifiProvider{}

type UnifiProvider struct {
	version string
}

type UnifiProviderModel struct {
	BaseURL  types.String `tfsdk:"base_url"`
	APIKey   types.String `tfsdk:"api_key"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Site     types.String `tfsdk:"site"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UnifiProvider{
			version: version,
		}
	}
}

func (p *UnifiProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "unifi"
	resp.Version = p.version
}

func (p *UnifiProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "The UniFi provider allows you to manage UniFi network infrastructure resources.",
		Attributes: map[string]schema.Attribute{
			"base_url": schema.StringAttribute{
				Description: "The base URL of the UniFi controller (e.g., https://192.168.1.1). " +
					"Can also be set via the UNIFI_BASE_URL environment variable.",
				Optional: true,
			},
			"api_key": schema.StringAttribute{
				Description: "API key for UniFi controller authentication (recommended). " +
					"This is the preferred authentication method. " +
					"Can also be set via the UNIFI_API_KEY environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"username": schema.StringAttribute{
				Description: "The username for UniFi controller authentication. " +
					"Only used if api_key is not provided. " +
					"Can also be set via the UNIFI_USERNAME environment variable.",
				Optional: true,
			},
			"password": schema.StringAttribute{
				Description: "The password for UniFi controller authentication. " +
					"Only used if api_key is not provided. " +
					"Can also be set via the UNIFI_PASSWORD environment variable.",
				Optional:  true,
				Sensitive: true,
			},
			"site": schema.StringAttribute{
				Description: "The UniFi site name. Defaults to 'default'. " +
					"Can also be set via the UNIFI_SITE environment variable.",
				Optional: true,
			},
			"insecure": schema.BoolAttribute{
				Description: "Skip TLS certificate verification. Defaults to false. " +
					"Can also be set via the UNIFI_INSECURE environment variable.",
				Optional: true,
			},
		},
	}
}

func (p *UnifiProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config UnifiProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use environment variables as fallbacks
	baseURL := os.Getenv("UNIFI_BASE_URL")
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	apiKey := os.Getenv("UNIFI_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	username := os.Getenv("UNIFI_USERNAME")
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	password := os.Getenv("UNIFI_PASSWORD")
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	site := os.Getenv("UNIFI_SITE")
	if !config.Site.IsNull() {
		site = config.Site.ValueString()
	}
	if site == "" {
		site = "default"
	}

	insecure := os.Getenv("UNIFI_INSECURE") == "true"
	if !config.Insecure.IsNull() {
		insecure = config.Insecure.ValueBool()
	}

	// Validate required configuration
	if baseURL == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("base_url"),
			"Missing UniFi Controller URL",
			"The provider cannot create the UniFi client as there is a missing or empty value for the UniFi controller URL. "+
				"Set the base_url value in the configuration or use the UNIFI_BASE_URL environment variable.",
		)
	}

	// Require either API key or username/password
	useAPIKey := apiKey != ""
	if !useAPIKey {
		if username == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("username"),
				"Missing Authentication Credentials",
				"The provider requires either an API key or username/password for authentication. "+
					"Set the api_key value (recommended) or both username and password in the configuration, "+
					"or use the UNIFI_API_KEY or UNIFI_USERNAME/UNIFI_PASSWORD environment variables.",
			)
		}

		if password == "" {
			resp.Diagnostics.AddAttributeError(
				path.Root("password"),
				"Missing Authentication Credentials",
				"The provider requires either an API key or username/password for authentication. "+
					"Set the api_key value (recommended) or both username and password in the configuration, "+
					"or use the UNIFI_API_KEY or UNIFI_USERNAME/UNIFI_PASSWORD environment variables.",
			)
		}
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create the UniFi client
	clientConfig := unifi.NetworkClientConfig{
		BaseURL:            baseURL,
		Site:               site,
		InsecureSkipVerify: insecure,
	}

	if useAPIKey {
		clientConfig.APIKey = apiKey
	} else {
		clientConfig.Username = username
		clientConfig.Password = password
	}

	client, err := unifi.NewNetworkClient(clientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create UniFi Client",
			"An unexpected error occurred when creating the UniFi client. "+
				"Error: "+err.Error(),
		)
		return
	}

	// Login to the controller (only needed for username/password auth)
	if !useAPIKey {
		if err := client.Login(ctx); err != nil {
			resp.Diagnostics.AddError(
				"Unable to Authenticate with UniFi Controller",
				"The provider failed to authenticate with the UniFi controller. "+
					"Please verify your credentials and controller URL. "+
					"Error: "+err.Error(),
			)
			return
		}
	}

	// Wrap client with auto-relogin capability
	wrappedClient := NewAutoLoginClient(client, clientConfig)

	// Make the client available to resources and data sources
	resp.DataSourceData = wrappedClient
	resp.ResourceData = wrappedClient
}

func (p *UnifiProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDevicePortOverrideResource,
		NewDynamicDNSResource,
		NewFirewallGroupResource,
		NewFirewallPolicyResource,
		NewFirewallRuleResource,
		NewFirewallZoneResource,
		NewNatRuleResource,
		NewNetworkResource,
		NewPortForwardResource,
		NewPortProfileResource,
		NewRADIUSProfileResource,
		NewStaticDNSResource,
		NewStaticRouteResource,
		NewTrafficRouteResource,
		NewTrafficRuleResource,
		NewUserGroupResource,
		NewUserResource,
		NewWLANResource,
	}
}

func (p *UnifiProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDeviceDataSource,
		NewDynamicDNSDataSource,
		NewFirewallGroupDataSource,
		NewFirewallPolicyDataSource,
		NewFirewallRuleDataSource,
		NewFirewallZoneDataSource,
		NewNatRuleDataSource,
		NewNetworkDataSource,
		NewPortForwardDataSource,
		NewPortProfileDataSource,
		NewRADIUSProfileDataSource,
		NewStaticDNSDataSource,
		NewStaticRouteDataSource,
		NewTrafficRouteDataSource,
		NewTrafficRuleDataSource,
		NewUserGroupDataSource,
		NewUserDataSource,
		NewWLANDataSource,
	}
}
