# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.10.2] - 2026-05-08

### Fixed

- `unifi_network` — removed the v0.8.2 plan-time mDNS validator, which was misreading the controller's settings model. The validator read `unifi_setting_usg.mdns_enabled` (a boolean field that the SDK declares but the v9 controller does not actually populate — Jackson silently dropped it). It would then conclude site-level mDNS was disabled and reject any per-network `mdns_enabled = true`, even on controllers where mDNS was actively configured and working via the actual `setting/mdns` settings key (`mode = auto`, `enabled_for_network_ids = [...]`). On v9 controllers this blocked **all** in-place changes to networks with `mdns_enabled = true`, including changes unrelated to mDNS — `lifecycle.ignore_changes` did not bypass it because `ModifyPlan` runs before `ignore_changes` is applied. The original "silent strip" footgun the validator was trying to prevent should be re-addressed once the SDK adds proper `setting/mdns` plumbing (tracked separately).

### Removed

- `unifi_network` — the `ResourceWithModifyPlan` interface assertion and the broken validator function. Removed the corresponding `TestAccNetworkResource_mdnsRequiresSiteLevel` test and the `testAccCheckSiteMDNSDisabled` / `testAccCheckSiteMDNSEnabled` test prechecks (they read the same wrong field). The `TestAccNetworkResource_mdnsHappyPathSiteLevelOn` test was renamed to `TestAccNetworkResource_mdnsRoundTrip` and now runs unconditionally.

## [0.10.1] - 2026-05-08

### Fixed

- `unifi_firewall_policy.schedule` and `unifi_traffic_rule.schedule` — `scheduleFromObject` / `trafficScheduleFromObject` helpers now check `IsUnknown()` on every string and set sub-attribute, not just `IsNull()`. With the schedule sub-attributes now `Optional+Computed` (v0.10.0), an interpolation that resolved to Unknown at plan time was being coerced to `""` via `ValueString()` and silently sent to the controller — clearing the field server-side or triggering controller validation errors.
- `unifi_traffic_rule` and `unifi_traffic_route` data sources — preserve the lookup `name` from config when the controller's GET endpoint returns an empty `name` (the controller drops `name` from every response shape; this matches the workaround the resources already had).

## [0.10.0] - 2026-05-07

This release bundles two cycles of probe-driven SDK corrections (v0.12.0 enum sweep + v0.13.0 traffic round-trip probe) into a single provider release. v0.12.0 was never tagged as a separate provider release.

### Breaking — schema attribute changes

- `unifi_firewall_policy.schedule.days_of_week` and `unifi_traffic_rule.schedule.days_of_week` — **renamed to `repeat_on_days`**. The previous attribute name and its uppercase day enum (`MONDAY`...`SUNDAY`) were Jackson-accepted but silently dropped by the controller (verified by the v0.13.0 SDK probe). New enum is the controller-confirmed lowercase 3-letter form: `[mon, tue, wed, thu, fri, sat, sun]`. **Hard-cut, no deprecation period** — any value previously sent for `days_of_week` was never persisted, so no working configuration breaks.
- `unifi_firewall_policy.schedule.mode` and `unifi_traffic_rule.schedule.mode` — enum widened from `[ALWAYS, CUSTOM]` to the controller-confirmed `[ALWAYS, EVERY_DAY, EVERY_WEEK, ONE_TIME_ONLY, CUSTOM]`. **Behavior change for `CUSTOM`**: the controller's `CUSTOM` mode actually requires `date_start` + `date_end` + `repeat_on_days`. The earlier provider/SDK semantics where `CUSTOM` meant "schedule on weekdays during business hours" should now use `EVERY_WEEK`. Configs using `mode = "CUSTOM"` without date fields will be rejected by the controller (`api.err.MissingDateRange`).
- `unifi_traffic_rule.network_id` (singular) — **renamed to `network_ids`** and changed from `String` to `Set of String`. The controller never persisted the single-value form (Jackson-accepted but dropped). Existing configs setting `network_id = X` must migrate to `network_ids = [X]`. Multi-network association is now expressible. *Stateful migration note*: Terraform users with state from older provider versions will see a "Unsupported argument" error on plan; they must rewrite their config before re-applying. State refresh will populate `network_ids` from the GET response.
- `unifi_traffic_route.fallback` — **removed**. The controller silently ignored this field (placebo); SDK v0.13.0 dropped the corresponding `Fallback` struct field. Configs setting `fallback = ...` will fail with "Unsupported argument" — but the value was never doing anything anyway.
- `unifi_firewall_policy` — `source.matching_target` and `destination.matching_target` enum **replaced** (from v0.12.0 cycle). The previous v0.9.0 enum was based on a single-mode controller probe that only revealed IP-style branch values. SDK v0.12.0's per-value probe discovered the actual enum is identity-aware: `[ANY, CLIENT, EXTERNAL_SOURCE, IID, IP, MAC, NETWORK, REGION, USER_IDENTITY, USER_IDENTITY_ONE_CLICK_VPN, USER_IDENTITY_ONE_CLICK_WIFI, VPN_USER]`. Configs that set `matching_target = "WEB" | "APP" | "APP_CATEGORY"` now fail at plan time. No working configuration breaks — those values were never accepted by the controller (v0.9.0 added them assuming controller-side carrier-field plumbing was forthcoming; the per-value probe proved they were dead).
- `unifi_nat_rule` — schema attributes `source_address`, `source_port`, `dest_address`, `dest_port`, `translated_ip`, `translated_port` **removed**. The UniFi v2 NAT API on current controllers responds "Unrecognized field" for each of these and returns HTTP 500 for any non-trivial create. The resource now manages only the rule shell (type, protocol, description, enabled, logging). Configure NAT translation via the UniFi UI until the upstream API stabilizes.
- `unifi_port_profile.poe_mode` — narrowed to `[auto, off]`. `pasv24` and `passthrough` are PortOverride-only.
- `unifi_port_profile.op_mode` — narrowed to `[switch]`. `mirror` and `aggregate` are PortOverride-only.
- `unifi_static_route.type` — narrowed to `[static-route]`. `interface-route` was never valid for this field (the per-route flavor lives in `static_route_type`).
- `unifi_setting_ips.advanced_filtering_preference` — `auto` removed.
- `unifi_setting_guest_access.auth` — `password` and `radius` removed; `facebook_wifi` added. Existing configs using `password` or `radius` must migrate to `hotspot`.
- `unifi_traffic_route.matching_target` — `APP` removed.

### Added

- `unifi_firewall_policy.schedule` and `unifi_traffic_rule.schedule` — new attributes `time_all_day` (Bool), `date_start` (String, YYYY-MM-DD), `date_end` (String, YYYY-MM-DD), `date` (String, YYYY-MM-DD) supporting the full `EVERY_DAY` / `EVERY_WEEK` / `ONE_TIME_ONLY` / `CUSTOM` mode set.
- `unifi_traffic_rule.network_ids` — set-of-string replacing the singular `network_id`. Enables multi-network rules.
- `unifi_firewall_policy` — `source.matching_target` and `destination.matching_target` accept identity-aware values: `CLIENT`, `EXTERNAL_SOURCE`, `MAC`, `USER_IDENTITY`, `USER_IDENTITY_ONE_CLICK_VPN`, `USER_IDENTITY_ONE_CLICK_WIFI`, `VPN_USER`. Carrier fields for these are not yet plumbed in the SDK; values pass validation but cannot produce a working policy until follow-up SDK PRs land.
- `unifi_traffic_rule.matching_target` — added `APP_CATEGORY` and `LOCAL_NETWORK`.

### Changed

- Bumped `unifi-go-sdk` from v0.11.0 to v0.13.0 (skipping a separate v0.12.0 provider release; both SDK cycles are folded into this one).
  - **v0.12.0**: probe sweep of every `isOneOf()` validator (17 enums corrected). Adds `TestEnumProbe` regression guard. Release notes: https://github.com/resnickio/unifi-go-sdk/releases/tag/v0.12.0
  - **v0.13.0**: round-trip probe of v2 traffic-* endpoints and `PolicySchedule`. Found 6 fields the controller silently dropped, flipped, or fabricated. Renames + struct shape corrections; transparent `name` re-injection on Create/Update return paths. Release notes: https://github.com/resnickio/unifi-go-sdk/releases/tag/v0.13.0
- `matchingTargetTypeFor` updated to drop dead v0.9.0 mappings (`WEB`/`APP`/`APP_CATEGORY` no longer in enum). Only `IP`/`NETWORK`/`REGION` and `IID` retain explicit mappings; identity-aware values fall through to empty-string default.

### Fixed

- `unifi_traffic_rule` and `unifi_traffic_route` — Create/Update no longer fail with `Provider produced inconsistent result after apply` on `.name` and `.schedule`. SDK v0.13.0 transparently re-injects `name` on Create/Update returns (the controller drops it from every response shape). The provider keeps a defensive `sdkToState` preservation of `name` from prior state — useful for the Read path, where the SDK falls back to LIST and still returns no name.
- `unifi_traffic_rule.schedule` — schedule values now round-trip correctly. The pre-v0.13.0 SDK silently corrupted the wire payload (`days_of_week` JSON tag mismatched the controller's `repeat_on_days`, day-name enum was wrong-cased). The provider's `repeat_on_days` attribute now reaches the controller intact.

### Known limitations (carried forward)

- Import of `unifi_traffic_rule` / `unifi_traffic_route` loses `name`. The controller's GET endpoints don't return `name` reliably; the SDK's `name` re-injection only covers Create/Update return paths. Tests use `ImportStateVerifyIgnore: ["name"]`; users importing must re-set `name` in config before re-applying. Documented in CLAUDE.md.
- Identity-aware `matching_target` values on `unifi_firewall_policy` pass validation but cannot produce a working policy (carrier fields not yet plumbed). Documented in CLAUDE.md.

### Background — what the probes taught us

v0.9.0 fixed a wrong enum based on a single Jackson deserialization error. The lesson v0.12.0 surfaced is that error messages only enumerate the values the deserializer actually evaluated against your input — so a single probe revealed only one branch of the real enum. The lesson v0.13.0 surfaced is even uglier: the controller silently *accepts* JSON tags it doesn't recognize (Jackson `@JsonIgnoreProperties(ignoreUnknown=true)`) and silently *flips* boolean fields when the JSON tag is wrong, with no error returned. Round-trip probing — write a value, read it back, check it's the same — is the only way to catch those. v0.13.0's `TestEnumProbe` and v0.12.0's per-value enum probe are now both durable regression guards.

## [0.9.1] - 2026-05-06

### Changed

- Bumped `unifi-go-sdk` from v0.10.0 to v0.11.0. The SDK release brings `PolicyEndpoint.Validate()` in line with the controller-confirmed `matching_target` enum that v0.9.0 corrected on the provider side: drops `DOMAIN`/`PORT_GROUP`/`ADDRESS_GROUP`, adds `APP`/`APP_CATEGORY`/`IID`/`WEB`. **No provider behavior change** — the v0.9.0 plan-time `OneOf` validator already enforced the corrected enum independent of the SDK. This bump removes the paper inconsistency between the two layers.

## [0.9.0] - 2026-05-06

### Breaking

- `unifi_firewall_policy` — `source.matching_target` and `destination.matching_target` no longer accept `'DOMAIN'`, `'PORT_GROUP'`, or `'ADDRESS_GROUP'`. **Those values were never accepted by the controller** — they were inherited from a Go SDK type definition that turned out to disagree with the controller's actual enum. Configs that included them have always failed at apply time with a Jackson deserialization error from the controller; this release moves the failure earlier in the pipeline (the schema validator now rejects them at plan time). No working configuration breaks. **Hard-cut, no deprecation period** — the values were dead on arrival, so a multi-release warn-only window would only delay the cleanup.

### Added

- `unifi_firewall_policy` — `source.matching_target` and `destination.matching_target` now accept `'APP'`, `'APP_CATEGORY'`, `'IID'`, and `'WEB'`. These are the matching modes the v9 controller actually exposes (full real enum: `[ANY, APP, APP_CATEGORY, IP, IID, NETWORK, REGION, WEB]`). **Carrier fields for these new modes are not yet implemented** in the Go SDK or provider — the values pass validation but cannot produce a working policy until follow-up SDK and provider PRs add the supporting fields (domain strings, app / category / object IDs). Tracked as a future minor release.
- `matching_target_type` derivation rules updated to mirror the corrected enum: `IP`, `NETWORK`, `REGION`, `WEB` → `SPECIFIC`; `APP`, `APP_CATEGORY`, `IID` → `OBJECT`.

### Background — how the wrong enum landed in v0.8.0

The Go SDK's `PolicyEndpoint.Validate()` listed `[ANY, IP, NETWORK, DOMAIN, REGION, PORT_GROUP, ADDRESS_GROUP]` as the valid set, derived from external documentation. v0.8.0 added a `OneOf` validator on the provider side that mirrored that list — which made plan-time validation tighter, but encoded the wrong data. A direct probe against the controller's `/proxy/network/v2/api/site/<site>/firewall-policies` endpoint with `matching_target: "DOMAIN"` returned this Jackson error:

> `Cannot deserialize value of type ... from String "DOMAIN": not one of the values accepted for Enum class: [APP, WEB, IP, APP_CATEGORY, NETWORK, IID, ANY, REGION]`

That single error message exposed the canonical enum. v0.9.0 adopts it.

## [0.8.2] - 2026-05-03

### Fixed

- `unifi_network` — setting `mdns_enabled = true` while site-wide gateway mDNS is disabled (`unifi_setting_usg.mdns_enabled = false`) now errors at plan time with a clear remediation message, instead of letting the controller silently strip the value at apply (which surfaced as "Provider produced inconsistent result after apply" with no useful explanation). Implementation: `ModifyPlan` fetches `SettingUSG` and validates the precondition. Same-apply caveat: if you are enabling site-level mDNS in the same apply as a network with `mdns_enabled = true`, you'll need to apply `unifi_setting_usg` first via `-target` or split into two applies.

## [0.8.1] - 2026-05-02

### Fixed

- `unifi_firewall_policy` — changing `action` from `ALLOW` to `BLOCK` (or `REJECT`) in a single apply no longer fails with `api.err.FirewallPolicyCreateRespondTrafficPolicyNotAllowed`. The provider now sends the `create_allow_respond` field and auto-derives it from `action` (`true` for ALLOW, `false` otherwise) so the wire payload stays consistent with the action change. Users who previously had to destroy + recreate to flip a policy's action can now edit in place.

### Added

- `unifi_firewall_policy.create_allow_respond` — `Optional + Computed` attribute exposing the controller's auto-respond toggle. Auto-derived from `action` when unset; explicit settings override.

### Known Issues

- `unifi_network` — site-wide `unifi_setting_usg.mdns_enabled` gates per-network `mdns_enabled = true`. A post-apply diagnostic was attempted in this release but discarded: the framework's "inconsistent result after apply" error fires before any provider-emitted warning surfaces. A future release will need a plan-time fetch of site state for a clean fix.

## [0.8.0] - 2026-05-02

### Fixed

- `unifi_firewall_policy` — **silent data loss when `matching_target` defaults to `ANY`**. The schema's static `"ANY"` default let users send `matching_target=ANY` alongside `ips=[...]`; the UniFi controller silently discarded `ips` and widened the policy to allow ANY destination in the zone. The `terraform plan` diff looked cosmetic (`matching_target: "IP" -> "ANY"`) so a careful operator could not tell the policy was about to widen. **Operators with imported policies should re-plan after upgrading; an unexpected `matching_target` change in the diff means the prior controller-side state had already been mutated.** ([#2](https://github.com/resnickio/terraform-provider-unifi/pull/2))
- `unifi_firewall_policy` — controller now requires `matching_target_type` (`SPECIFIC` for inline values, `OBJECT` for group references) on every IP/NETWORK match. The provider was not sending this field, causing `api.err.MissingFirewallDestinationMatchingTargetType` from v9 controllers. The provider now derives it transparently from `matching_target`. ([#2](https://github.com/resnickio/terraform-provider-unifi/pull/2))
- `unifi_network` — `mdns_enabled` and `upnp_lan_enabled` were marked `Optional` only with no default, so omitting them from config caused `Provider produced inconsistent result after apply: was null, but now cty.False` whenever the controller returned the field. Both attributes are now `Optional + Computed` and resolve to whatever the controller persists. ([#3](https://github.com/resnickio/terraform-provider-unifi/pull/3))

### Changed

- `unifi_firewall_policy` — `source.matching_target` and `destination.matching_target` are now auto-derived from sibling fields when the user leaves them unset: `ips` non-empty → `"IP"`, `network_id` non-empty → `"NETWORK"`, otherwise `"ANY"`. Users who previously specified `matching_target` explicitly are unaffected; users who left it unset will see plans correctly resolve to `IP`/`NETWORK` instead of the broken `ANY` default.
- `unifi_firewall_policy` — `matching_target` now validates against the enum at plan time via `stringvalidator.OneOf(...)`. Typos that previously reached the controller now fail in `terraform plan` with a clear error.
- `unifi_firewall_policy` — explicit `matching_target = "ANY"` alongside `ips` or `network_id` is now rejected at plan time as a config error, since UniFi silently strips those fields under `matching_target=ANY`.

### Known Issues

- `unifi_firewall_policy` — matching targets `DOMAIN`, `REGION`, `PORT_GROUP`, `ADDRESS_GROUP` are not usable. The Go SDK's `PolicyEndpoint` type lacks fields to carry the corresponding match data (domains, regions, group IDs); the provider can set the marker but not the values. Manage these policies via the UniFi UI for now.
- `unifi_firewall_policy` — changing `action` from `ALLOW` to `BLOCK` in a single apply can fail on UniFi v9 with `api.err.FirewallPolicyCreateRespondTrafficPolicyNotAllowed`. Workaround: destroy and recreate the policy. Pre-existing; not introduced by 0.8.0.
- `unifi_network` — site-wide `unifi_setting_usg.mdns_enabled` gates whether per-network `mdns_enabled = true` takes effect. The controller silently strips per-network `true` if the site-level toggle is off, with no error.

## [0.7.2] - 2026-03-25

### Fixed

- `unifi_firewall_policy` — `index` made `Computed`-only. The controller auto-reassigns it on every create, so any user-set value caused `inconsistent result after apply` or invalid plans. Removes the `serverValueOnCreateModifier` plan modifier added in v0.7.1 (which the framework rejected because it can't return Unknown when config has a concrete value). Fixes #1.

## [0.7.1] - 2026-03-25

### Fixed

- `unifi_firewall_policy` — `index` field caused `Provider produced inconsistent result after apply` when the controller assigned a value different from the user's plan. Initial fix attempt; superseded by v0.7.2's clean read-only approach.

## [0.7.0] - 2026-03-02

### Added

#### New Resources
- `unifi_account` - RADIUS accounts for 802.1X / VPN authentication
- `unifi_content_filtering` - Content filtering settings (singleton)
- `unifi_device` - Manage already-adopted device settings (radio overrides, SNMP, etc.)
- `unifi_setting_guest_access` - Guest portal / captive portal settings (singleton)
- `unifi_setting_ips` - IPS/IDS and threat management settings (singleton)
- `unifi_setting_magic_site_to_site_vpn` - Magic site-to-site VPN settings (singleton)
- `unifi_setting_mgmt` - Site management settings (singleton: auto-upgrade, LED, SSH, alerts)
- `unifi_setting_radius` - Site RADIUS server settings (singleton)
- `unifi_setting_snmp` - SNMP monitoring settings (singleton)
- `unifi_setting_teleport` - Teleport VPN settings (singleton)
- `unifi_setting_usg` - Site USG/gateway settings (singleton: UPnP, mDNS, offloading)
- `unifi_site` - Manage controller sites

#### New Data Sources
- `unifi_account` - Look up RADIUS account by ID or name
- `unifi_acl_rule` - Look up ACL rule by ID or name (read-only)
- `unifi_active_client` - Look up active (connected) client by MAC or display name
- `unifi_admin` - Look up controller admin by ID or name
- `unifi_ap_group` - Look up AP group by ID or name
- `unifi_backup` - List all controller backups
- `unifi_content_filtering` - Read current content filtering configuration
- `unifi_qos_rule` - Look up QoS rule by ID or name
- `unifi_site` - Look up site by ID or name
- `unifi_vpn_connection` - Look up VPN connection by ID or name
- `unifi_wan_sla` - Look up WAN SLA monitor by ID or name

### Fixed

- IPS settings: corrected setting key, handle honeypot / ad-blocking fields.
- Traffic rules: added missing `schedule` nested attribute and `app_category_ids`.
- Port profiles: added missing `dot1x_idle_timeout` and `egress_rate_limit` fields.
- Teleport: `enabled` now correctly handled as a `*bool`.
- IPv6: RA lifetime fields use the SDK's `FlexInt` for string-or-int unmarshalling.

### Changed

- Updated `unifi-go-sdk` to v0.10.0 (`FlexInt` support).
- Added `flexIntValueOrNull` helper in `utils.go`.

## [0.6.0] - 2026-02-27

### Added

- `unifi_network` — IPv6 support via a new nested `ipv6` attribute (21 fields covering interface type, prefix delegation, RA, DHCPv6, SLAAC) with full acceptance test coverage.

### Fixed

- `unifi_network` — `dhcp_dns_enabled` plan/state mismatch when DNS servers auto-enable it.
- Build: legacy `GNUmakefile` was shadowing `Makefile`, breaking the `testacc-run` target. Removed.

### Changed

- **Breaking:** the previous flat `ipv6_setting_preference` attribute is replaced by the nested `ipv6` block. Operators using IPv6 must migrate their configurations.

## [0.5.0] - 2026-02-23

### Added

- `unifi_network` — `dhcp_dns_enabled` and `dhcp_tftp_server` attributes.

### Changed

- Sweeper expanded; CLAUDE.md updated; pinned `terraform-plugin-docs` version.

## [0.4.0] - 2026-02-14

### Added

#### New Resources
- `unifi_user` - Client device records with DHCP reservations, fixed IPs, device naming, local DNS records, blocking, and user group assignments

#### New Data Sources
- `unifi_user` - Look up user by ID or MAC address

### Changed
- Updated unifi-go-sdk from v0.5.0 to v0.6.0 (adds User CRUD support)
- Provider now registers 17 resources and 18 data sources

## [0.3.1] - 2026-01-21

### Fixed

- `unifi_device_port_override` — race condition when multiple port overrides on the same device used `for_each`. Read-modify-write operations could overwrite each other; added per-device locking to serialize updates.

## [0.3.0] - 2026-01-21

### Added

- `unifi_device` data source — look up devices by MAC or name.
- `unifi_device_port_override` resource — configure per-port settings on switches: profile assignment, PoE mode, port name, VLANs per port, link aggregation, port isolation, rate limiting. Import/export uses `device_id:port_idx` format.

### Changed

- Updated `unifi-go-sdk` to v0.5.0 (device management support).

## [0.2.3] - 2026-01-18

### Fixed

- `unifi_static_dns` and `unifi_network` — state inconsistency between plan and apply.
- Network resource documentation regenerated to match the corrected schema.

## [0.2.2] - 2026-01-18

### Added

- GoReleaser config and GitHub Actions release workflow — first release published to the Terraform Registry.

## [0.2.1] - 2026-01-18

### Fixed

- `unifi_port_profile` - Corrected `tagged_vlan_mgmt` valid values to "auto", "block_all", "custom" (was incorrectly "all", "block", "custom")
- Test fixes for dynamic DNS (use "custom" service which is universally supported)
- Test fixes for traffic rules (added required `matching_target` and `target_devices` fields)
- Test fixes for WLAN data source (removed reference to non-existent `unifi_ap_group` data source)
- Test fixes for firewall rules (added precheck to skip on zone-based firewall controllers)

### Known Issues

- `unifi_traffic_rule` - Provider has state consistency issues with default values for `bandwidth_limit` and `schedule`
- `unifi_traffic_route` - Similar state consistency issues
- `unifi_static_dns`, `unifi_nat_rule`, `unifi_radius_profile` - Schema issues with nested blocks vs attributes

## [0.2.0] - 2026-01-16

### Added

#### New Resources
- `unifi_static_dns` - Static DNS records (v2 API)
- `unifi_dynamic_dns` - Dynamic DNS configuration
- `unifi_nat_rule` - NAT rules (v2 API)
- `unifi_traffic_rule` - Traffic rules for QoS and blocking (v2 API)
- `unifi_traffic_route` - Traffic routes for policy-based routing (v2 API)
- `unifi_radius_profile` - RADIUS authentication profiles
- `unifi_port_profile` - Switch port profiles with VLAN, PoE, 802.1X, and storm control

#### New Data Sources
- `unifi_static_dns` - Look up static DNS by ID or key
- `unifi_dynamic_dns` - Look up dynamic DNS by ID or hostname
- `unifi_nat_rule` - Look up NAT rule by ID or description
- `unifi_traffic_rule` - Look up traffic rule by ID or description
- `unifi_traffic_route` - Look up traffic route by ID or description
- `unifi_port_profile` - Look up port profile by ID or name
- `unifi_port_forward` - Look up port forward by ID or name
- `unifi_firewall_policy` - Look up firewall policy by ID or name
- `unifi_firewall_rule` - Look up firewall rule by ID or name
- `unifi_static_route` - Look up static route by ID or name
- `unifi_user_group` - Look up user group by ID or name
- `unifi_wlan` - Look up WLAN by ID or name

#### Documentation
- Added example configurations for all data sources
- Expanded CLAUDE.md with write-only fields pattern and data source pattern
- Updated README.md with complete resource and data source listing

### Fixed
- WLAN passphrase now correctly preserved during Update operations (previously caused state drift)

### Changed
- Provider now registers 16 resources and 15 data sources (up from 9 resources and 3 data sources)

## [0.1.1] - 2026-01-15

### Added
- `unifi_firewall_zone` data source
- `unifi_firewall_group` data source

### Fixed
- Documentation improvements

## [0.1.0] - 2026-01-01

### Added

#### Resources
- `unifi_network` - VLAN networks with DHCP configuration
- `unifi_firewall_group` - Address and port groups
- `unifi_firewall_rule` - Legacy firewall rules
- `unifi_firewall_policy` - Zone-based firewall policies (v2 API)
- `unifi_firewall_zone` - Firewall zones (v2 API)
- `unifi_port_forward` - Port forwarding rules
- `unifi_static_route` - Static routing
- `unifi_user_group` - User/bandwidth groups
- `unifi_wlan` - Wireless networks (SSIDs)

#### Data Sources
- `unifi_network` - Look up network by ID or name

#### Infrastructure
- Auto-relogin client with retry logic and rate limiting
- Support for both API key and username/password authentication
- Configurable timeouts for all CRUD operations
- Comprehensive acceptance test suite with sweepers
