package cli

import "github.com/spf13/cobra"

func OptionalBool(cmd *cobra.Command, name string) *bool {
	if !cmd.Flags().Changed(name) {
		return nil
	}

	value, _ := cmd.Flags().GetBool(name)

	return &value
}

func OptionalInt(cmd *cobra.Command, name string) *int {
	if !cmd.Flags().Changed(name) {
		return nil
	}

	value, _ := cmd.Flags().GetInt(name)

	return &value
}

func OptionalFloat(cmd *cobra.Command, name string) *float64 {
	if !cmd.Flags().Changed(name) {
		return nil
	}

	value, _ := cmd.Flags().GetFloat64(name)

	return &value
}

func OptionalString(cmd *cobra.Command, name string) *string {
	if !cmd.Flags().Changed(name) {
		return nil
	}

	value, _ := cmd.Flags().GetString(name)

	return &value
}
