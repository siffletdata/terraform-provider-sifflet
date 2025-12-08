package domain

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestDomainResourceStateUpgradeV0(t *testing.T) {
	ctx := context.Background()
	r := &domainResource{}
	upgraders := r.UpgradeState(ctx)
	upgraderV0 := upgraders[0]
	priorSchema := upgraderV0.PriorSchema

	t.Run("upgrade_schema_uris_list_to_set", func(t *testing.T) {
		rawStateValue := tftypes.NewValue(
			priorSchema.Type().TerraformType(ctx),
			map[string]tftypes.Value{
				"id":          tftypes.NewValue(tftypes.String, "00000000-0000-0000-0000-000000000001"),
				"name":        tftypes.NewValue(tftypes.String, "Test Domain"),
				"description": tftypes.NewValue(tftypes.String, nil),
				"dynamic_content_definition": tftypes.NewValue(
					tftypes.Object{
						AttributeTypes: map[string]tftypes.Type{
							"logical_operator": tftypes.String,
							"conditions": tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"logical_operator": tftypes.String,
										"schema_uris":      tftypes.List{ElementType: tftypes.String},
										"tags":             tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String, "name": tftypes.String, "kind": tftypes.String}}},
									},
								},
							},
						},
					},
					map[string]tftypes.Value{
						"logical_operator": tftypes.NewValue(tftypes.String, "AND"),
						"conditions": tftypes.NewValue(
							tftypes.List{
								ElementType: tftypes.Object{
									AttributeTypes: map[string]tftypes.Type{
										"logical_operator": tftypes.String,
										"schema_uris":      tftypes.List{ElementType: tftypes.String},
										"tags":             tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String, "name": tftypes.String, "kind": tftypes.String}}},
									},
								},
							},
							[]tftypes.Value{
								tftypes.NewValue(
									tftypes.Object{
										AttributeTypes: map[string]tftypes.Type{
											"logical_operator": tftypes.String,
											"schema_uris":      tftypes.List{ElementType: tftypes.String},
											"tags":             tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String, "name": tftypes.String, "kind": tftypes.String}}},
										},
									},
									map[string]tftypes.Value{
										"logical_operator": tftypes.NewValue(tftypes.String, "IS"),
										"schema_uris": tftypes.NewValue(
											tftypes.List{ElementType: tftypes.String},
											[]tftypes.Value{
												tftypes.NewValue(tftypes.String, "bigquery://project/dataset"),
												tftypes.NewValue(tftypes.String, "snowflake://account/database"),
											},
										),
										"tags": tftypes.NewValue(tftypes.List{ElementType: tftypes.Object{AttributeTypes: map[string]tftypes.Type{"id": tftypes.String, "name": tftypes.String, "kind": tftypes.String}}}, nil),
									},
								),
							},
						),
					},
				),
				"static_content_definition": tftypes.NewValue(
					tftypes.Object{AttributeTypes: map[string]tftypes.Type{"asset_uris": tftypes.Set{ElementType: tftypes.String}}},
					nil,
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
				Schema: domainResourceSchema(),
			},
		}

		upgraderV0.StateUpgrader(ctx, req, resp)

		if resp.Diagnostics.HasError() {
			t.Fatalf("Unexpected error: %v", resp.Diagnostics)
		}

		var upgradedState domainModel
		resp.State.Get(ctx, &upgradedState)

		// Verify schema_uris was converted from List to Set
		var dynamicDef dynamicContentDefinitionModel
		upgradedState.DynamicContentDefinition.As(ctx, &dynamicDef, basetypes.ObjectAsOptions{})

		var conditions []types.Object
		dynamicDef.Conditions.ElementsAs(ctx, &conditions, false)

		var condition dynamicContentDefinitionConditionModel
		conditions[0].As(ctx, &condition, basetypes.ObjectAsOptions{})

		var schemaUris []types.String
		condition.SchemaUris.ElementsAs(ctx, &schemaUris, false)

		if len(schemaUris) != 2 {
			t.Errorf("Expected 2 schema_uris, got %d", len(schemaUris))
		}

		// Verify values are preserved
		schemaUrisMap := make(map[string]bool)
		for _, uri := range schemaUris {
			schemaUrisMap[uri.ValueString()] = true
		}

		if !schemaUrisMap["bigquery://project/dataset"] || !schemaUrisMap["snowflake://account/database"] {
			t.Error("schema_uris values not preserved correctly")
		}
	})
}

func TestDomainResourceSchemaVersion(t *testing.T) {
	schema := domainResourceSchema()
	if schema.Version != 1 {
		t.Errorf("Expected schema version 1, got %d", schema.Version)
	}
}

func TestDomainResourceImplementsUpgradeState(t *testing.T) {
	var _ resource.ResourceWithUpgradeState = &domainResource{}
}
