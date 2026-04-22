package user

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var permissionTftypeObject = tftypes.Object{
	AttributeTypes: map[string]tftypes.Type{
		"domain_id":   tftypes.String,
		"domain_role": tftypes.String,
	},
}

func TestUserResourceStateUpgradeV0(t *testing.T) {
	ctx := context.Background()
	r := &userResource{}
	upgraders := r.UpgradeState(ctx)
	upgraderV0 := upgraders[0]
	priorSchema := upgraderV0.PriorSchema

	t.Run("upgrade_permissions_list_to_set", func(t *testing.T) {
		rawStateValue := tftypes.NewValue(
			priorSchema.Type().TerraformType(ctx),
			map[string]tftypes.Value{
				"id":    tftypes.NewValue(tftypes.String, "00000000-0000-0000-0000-000000000001"),
				"name":  tftypes.NewValue(tftypes.String, "Test User"),
				"email": tftypes.NewValue(tftypes.String, "test@example.com"),
				"role":  tftypes.NewValue(tftypes.String, "EDITOR"),
				"permissions": tftypes.NewValue(
					tftypes.List{ElementType: permissionTftypeObject},
					[]tftypes.Value{
						tftypes.NewValue(permissionTftypeObject, map[string]tftypes.Value{
							"domain_id":   tftypes.NewValue(tftypes.String, "aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb"),
							"domain_role": tftypes.NewValue(tftypes.String, "VIEWER"),
						}),
						tftypes.NewValue(permissionTftypeObject, map[string]tftypes.Value{
							"domain_id":   tftypes.NewValue(tftypes.String, "ccccdddd-cccc-dddd-cccc-ddddccccdddd"),
							"domain_role": tftypes.NewValue(tftypes.String, "EDITOR"),
						}),
					},
				),
				"auth_types": tftypes.NewValue(
					tftypes.Set{ElementType: tftypes.String},
					[]tftypes.Value{
						tftypes.NewValue(tftypes.String, "SAML2"),
					},
				),
			},
		)

		req := resource.UpgradeStateRequest{
			State: &tfsdk.State{
				Raw:    rawStateValue,
				Schema: *priorSchema,
			},
		}

		resp := &resource.UpgradeStateResponse{
			State: tfsdk.State{
				Schema: userResourceSchema(),
			},
		}

		upgraderV0.StateUpgrader(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		var upgradedState userModel
		resp.State.Get(ctx, &upgradedState)

		// Verify other fields are preserved
		if upgradedState.Id.ValueString() != "00000000-0000-0000-0000-000000000001" {
			t.Errorf("Expected id %q, got %q", "00000000-0000-0000-0000-000000000001", upgradedState.Id.ValueString())
		}
		if upgradedState.Email.ValueString() != "test@example.com" {
			t.Errorf("Expected email %q, got %q", "test@example.com", upgradedState.Email.ValueString())
		}
		if upgradedState.Role.ValueString() != "EDITOR" {
			t.Errorf("Expected role %q, got %q", "EDITOR", upgradedState.Role.ValueString())
		}

		// Verify permissions is not null and contains the right elements
		if upgradedState.Permissions.IsNull() {
			t.Fatal("permissions should not be null after upgrade")
		}

		var permissions []permissionModel
		upgradedState.Permissions.ElementsAs(ctx, &permissions, false)

		if len(permissions) != 2 {
			t.Fatalf("Expected 2 permissions, got %d", len(permissions))
		}

		// Verify all permissions are preserved (as a set, order is not guaranteed)
		expectedPermissions := map[string]string{
			"aaaabbbb-aaaa-bbbb-aaaa-bbbbaaaabbbb": "VIEWER",
			"ccccdddd-cccc-dddd-cccc-ddddccccdddd": "EDITOR",
		}
		for _, p := range permissions {
			expectedRole, ok := expectedPermissions[p.DomainId.ValueString()]
			if !ok {
				t.Errorf("Unexpected domain_id %q in upgraded permissions", p.DomainId.ValueString())
				continue
			}
			if p.DomainRole.ValueString() != expectedRole {
				t.Errorf("Expected domain_role %q for domain %q, got %q", expectedRole, p.DomainId.ValueString(), p.DomainRole.ValueString())
			}
		}
	})
}

func TestUserResourceSchemaVersion(t *testing.T) {
	schema := userResourceSchema()
	if schema.Version != 1 {
		t.Errorf("Expected schema version 1, got %d", schema.Version)
	}
}

func TestUserResourceImplementsUpgradeState(t *testing.T) {
	var _ resource.ResourceWithUpgradeState = &userResource{}
}
