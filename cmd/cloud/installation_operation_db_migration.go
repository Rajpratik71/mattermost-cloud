// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.
//

package main

import (
	"github.com/mattermost/mattermost-cloud/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func init() {
	installationDBMigrationRequestCmd.Flags().String("installation", "", "The id of the installation to be migrated.")
	installationDBMigrationRequestCmd.Flags().String("destination-db", model.InstallationDatabaseMultiTenantRDSPostgres, "The destination database type.")
	installationDBMigrationRequestCmd.Flags().String("multi-tenant-db", "", "The id of the destination multi tenant db.")
	installationDBMigrationRequestCmd.MarkFlagRequired("installation")
	installationDBMigrationRequestCmd.MarkFlagRequired("multi-tenant-db")

	installationDBMigrationsListCmd.Flags().String("installation", "", "The id of the installation to query operations.")
	installationDBMigrationsListCmd.Flags().String("state", "", "The state to filter operations by.")
	registerTableOutputFlags(installationDBMigrationsListCmd)
	registerPagingFlags(installationDBMigrationsListCmd)

	installationDBMigrationGetCmd.Flags().String("db-migration", "", "The id of the installation db migration operation.")
	installationDBMigrationGetCmd.MarkFlagRequired("db-migration")

	installationDBMigrationCommitCmd.Flags().String("db-migration", "", "The id of the installation db migration operation.")
	installationDBMigrationCommitCmd.MarkFlagRequired("db-migration")

	installationDBMigrationRollbackCmd.Flags().String("db-migration", "", "The id of the installation db migration operation.")
	installationDBMigrationRollbackCmd.MarkFlagRequired("db-migration")

	installationDBMigrationOperationCmd.AddCommand(installationDBMigrationRequestCmd)
	installationDBMigrationOperationCmd.AddCommand(installationDBMigrationsListCmd)
	installationDBMigrationOperationCmd.AddCommand(installationDBMigrationGetCmd)
	installationDBMigrationOperationCmd.AddCommand(installationDBMigrationCommitCmd)
	installationDBMigrationOperationCmd.AddCommand(installationDBMigrationRollbackCmd)
}

var installationDBMigrationOperationCmd = &cobra.Command{
	Use:   "db-migration",
	Short: "Manipulate installation db migration operations managed by the provisioning server.",
}

var installationDBMigrationRequestCmd = &cobra.Command{
	Use:   "request",
	Short: "Request database migration to different DB",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		// For now only multi-tenant postgres DB is supported.
		installationID, _ := command.Flags().GetString("installation")
		destinationDB, _ := command.Flags().GetString("destination-db")
		multiTenantDBID, _ := command.Flags().GetString("multi-tenant-db")

		request := &model.InstallationDBMigrationRequest{
			InstallationID:         installationID,
			DestinationDatabase:    destinationDB,
			DestinationMultiTenant: &model.MultiTenantDBMigrationData{DatabaseID: multiTenantDBID},
		}

		dryRun, _ := command.Flags().GetBool("dry-run")
		if dryRun {
			err := printJSON(request)
			if err != nil {
				return errors.Wrap(err, "failed to print API request")
			}

			return nil
		}

		migrationOperation, err := client.MigrateInstallationDatabase(request)
		if err != nil {
			return errors.Wrap(err, "failed to request installation database migration")
		}

		err = printJSON(migrationOperation)
		if err != nil {
			return err
		}

		return nil
	},
}

var installationDBMigrationsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installation database migration operations",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		installationID, _ := command.Flags().GetString("installation")
		state, _ := command.Flags().GetString("state")
		paging := parsePagingFlags(command)

		request := &model.GetInstallationDBMigrationOperationsRequest{
			Paging:         paging,
			InstallationID: installationID,
			State:          state,
		}

		dbMigrationOperations, err := client.GetInstallationDBMigrationOperations(request)
		if err != nil {
			return errors.Wrap(err, "failed to list installation database migration operations")
		}

		if enabled, customCols := tableOutputEnabled(command); enabled {
			var keys []string
			var vals [][]string

			if len(customCols) > 0 {
				data := make([]interface{}, 0, len(dbMigrationOperations))
				for _, elem := range dbMigrationOperations {
					data = append(data, elem)
				}
				keys, vals, err = prepareTableData(customCols, data)
				if err != nil {
					return errors.Wrap(err, "failed to prepare table output")
				}
			} else {
				keys, vals = defaultDBMigrationOperationTableData(dbMigrationOperations)
			}

			printTable(keys, vals)
			return nil
		}

		err = printJSON(dbMigrationOperations)
		if err != nil {
			return err
		}

		return nil
	},
}

func defaultDBMigrationOperationTableData(ops []*model.InstallationDBMigrationOperation) ([]string, [][]string) {
	keys := []string{"ID", "INSTALLATION ID", "STATE", "REQUEST AT"}
	vals := make([][]string, 0, len(ops))

	for _, migration := range ops {
		vals = append(vals, []string{
			migration.ID,
			migration.InstallationID,
			string(migration.State),
			model.TimeFromMillis(migration.RequestAt).Format("2006-01-02 15:04:05 -0700 MST"),
		})
	}
	return keys, vals
}

var installationDBMigrationGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Fetches given installation database migration operation.",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		dbMigrationID, _ := command.Flags().GetString("db-migration")

		migrationOperation, err := client.GetInstallationDBMigrationOperation(dbMigrationID)
		if err != nil {
			return errors.Wrap(err, "failed to get installation database migration")
		}

		err = printJSON(migrationOperation)
		if err != nil {
			return err
		}

		return nil
	},
}

var installationDBMigrationCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Commits database migration",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		dbMigrationID, _ := command.Flags().GetString("db-migration")

		migrationOperation, err := client.CommitInstallationDBMigration(dbMigrationID)
		if err != nil {
			return errors.Wrap(err, "failed to commit installation database migration")
		}

		err = printJSON(migrationOperation)
		if err != nil {
			return err
		}

		return nil
	},
}

var installationDBMigrationRollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Triggers rollback of database migration",
	RunE: func(command *cobra.Command, args []string) error {
		command.SilenceUsage = true

		serverAddress, _ := command.Flags().GetString("server")
		client := model.NewClient(serverAddress)

		dbMigrationID, _ := command.Flags().GetString("db-migration")

		migrationOperation, err := client.RollbackInstallationDBMigration(dbMigrationID)
		if err != nil {
			return errors.Wrap(err, "failed to trigger rollback of installation database migration")
		}

		err = printJSON(migrationOperation)
		if err != nil {
			return err
		}

		return nil
	},
}
