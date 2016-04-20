package database

import (
	"fmt"

	"github.com/solher/snakepit/root"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var Cmd = &cobra.Command{
	Use:     "database",
	Aliases: []string{"db"},
	Short:   "Database management",
}

func init() {
	Cmd.AddCommand(create)
	Cmd.AddCommand(migrate)
	Cmd.AddCommand(drop)
	Cmd.AddCommand(seed)
	Cmd.AddCommand(reset)
}

var Create, Migrate, Drop, Seed func(v *viper.Viper) error

var create = &cobra.Command{
	Use:   "create",
	Short: "Creates the app database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Creating database...")

		if err := Create(root.Viper); err != nil {
			return err
		}

		fmt.Println("Done.")

		return nil
	},
}

var migrate = &cobra.Command{
	Use:   "migrate",
	Short: "Creates the app collections in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Migrating database...")

		if err := Migrate(root.Viper); err != nil {
			return err
		}

		fmt.Println("Done.")

		return nil
	},
}

var seed = &cobra.Command{
	Use:   "seed",
	Short: "Synchronizes the local and distant seeds",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Seeding database...")

		if err := Seed(root.Viper); err != nil {
			return err
		}

		fmt.Println("Done.")

		return nil
	},
}

var drop = &cobra.Command{
	Use:   "drop",
	Short: "Drops the app database",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Dropping database...")

		if err := Drop(root.Viper); err != nil {
			return err
		}

		fmt.Println("Done.")

		return nil
	},
}

var reset = &cobra.Command{
	Use:   "reset",
	Short: "Alias for drop, create, migrate, seed",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Resetting database...")

		if err := Drop(root.Viper); err != nil {
			return err
		}

		if err := Create(root.Viper); err != nil {
			return err
		}

		if err := Migrate(root.Viper); err != nil {
			return err
		}

		if err := Seed(root.Viper); err != nil {
			return err
		}

		fmt.Println("Done.")

		return nil
	},
}
