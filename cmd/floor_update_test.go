package cmd

import "testing"

func TestFloorUpdateAliasesSchema(t *testing.T) {
	cmd, err := resolveSchemaTarget([]string{"floor", "update"})
	if err != nil {
		t.Fatal(err)
	}
	flag := cmd.Flags().Lookup("aliases")
	if flag == nil || flag.Value.Type() != "stringSlice" {
		t.Fatal("missing repeatable aliases flag")
	}
	schema := buildSchemaTree(cmd)
	for _, variant := range schema.OutputContract.Variants {
		found := false
		for _, field := range variant.Data.Fields {
			if field.Name == "changes" {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing diff contract in %s", variant.Name)
		}
	}
}
