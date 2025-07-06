// Package cli provides common utilities and helpers for CLI commands.
//
// This package contains shared functionality used across multiple CLI commands,
// including configuration loading, repository initialization, input validation,
// and error handling.
//
// Key Components:
//
//   - Context: Encapsulates common dependencies (config, repository) for commands
//   - Validation: Input validation functions for task fields
//   - Error handling: Consistent error formatting and exit codes
//
// Usage:
//
// Commands should use LoadContext to initialize their dependencies:
//
//	func runCommand(cmd *cobra.Command, args []string) error {
//	    ctx, err := cli.LoadContext(cmd)
//	    if err != nil {
//	        return err
//	    }
//	    // Use ctx.Config and ctx.Repo
//	}
//
// Input validation example:
//
//	taskID, err := cli.NormalizeTaskID(args[0])
//	if err != nil {
//	    return err
//	}
//	
//	status, err := cli.ValidateStatus(statusStr)
//	if err != nil {
//	    return err
//	}
package cli