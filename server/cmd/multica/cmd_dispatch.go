package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// dispatchCmd is the parent command for dispatch contract operations.
// Usage: multica dispatch [create|list|show|cancel]
var dispatchCmd = &cobra.Command{
	Use:   "dispatch",
	Short: "Manage dispatch contracts between issues",
	Long: `Dispatch contracts create a formal notification link between two issues.
When trigger events fire on the source issue, the target issue is notified.

Contracts are distinct from parent-child relationships — they represent
temporary, event-driven coordination agreements, e.g. "notify issue X
when a status change occurs on issue Y."`,
}

// dispatchCreateCmd: multica dispatch create --from <issue-id> --to <issue-id>
// --on <events> [--target <issue-id>]
var dispatchCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a dispatch contract between two issues",
	Long: `Create a dispatch contract that links two issues for cross-issue notification.

When any of the trigger events (--on) fire on the source issue (--from),
the target issue (--to or --target) is notified.

Examples:
  multica dispatch create --from OXY-123 --to OXY-456 --on status_changed
  multica dispatch create --from OXY-123 --to OXY-789 --on "status_changed,comment:created" --target OXY-456`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient(cmd)
		if err != nil {
			return err
		}

		fromIssueID, _ := cmd.Flags().GetString("from")
		toIssueID, _ := cmd.Flags().GetString("to")
		onEvents, _ := cmd.Flags().GetString("on")
		targetIssue, _ := cmd.Flags().GetString("target")

		if fromIssueID == "" || toIssueID == "" {
			return fmt.Errorf("--from and --to are required")
		}
		if onEvents == "" {
			return fmt.Errorf("--on (trigger events) is required")
		}

		events := strings.Split(onEvents, ",")
		for i := range events {
			events[i] = strings.TrimSpace(events[i])
		}

		body := map[string]any{
			"from_issue_id":  fromIssueID,
			"to_issue_id":    toIssueID,
			"trigger_events": events,
		}
		if targetIssue != "" {
			body["target_issue"] = targetIssue
		}

		var result map[string]any
		if err := client.PostJSON(cmd.Context(), "/api/dispatch/contracts", body, &result); err != nil {
			return fmt.Errorf("create dispatch contract: %w", err)
		}

		printDispatchJSON(result)
		return nil
	},
}

// dispatchListCmd: multica dispatch list [--status pending|triggered|fulfilled|cancelled]
var dispatchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dispatch contracts in the workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient(cmd)
		if err != nil {
			return err
		}

		params := url.Values{}
		if status, _ := cmd.Flags().GetString("status"); status != "" {
			params.Set("status", status)
		}

		path := "/api/dispatch/contracts"
		if q := params.Encode(); q != "" {
			path += "?" + q
		}

		var result map[string]any
		if err := client.GetJSON(cmd.Context(), path, &result); err != nil {
			return fmt.Errorf("list dispatch contracts: %w", err)
		}

		printDispatchJSON(result)
		return nil
	},
}

// dispatchShowCmd: multica dispatch show <contract-id>
var dispatchShowCmd = &cobra.Command{
	Use:   "show <contract-id>",
	Short: "Show dispatch contract details",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient(cmd)
		if err != nil {
			return err
		}

		path := "/api/dispatch/contracts/" + url.PathEscape(args[0])
		var result map[string]any
		if err := client.GetJSON(cmd.Context(), path, &result); err != nil {
			return fmt.Errorf("get dispatch contract: %w", err)
		}

		printDispatchJSON(result)
		return nil
	},
}

// dispatchCancelCmd: multica dispatch cancel <contract-id>
var dispatchCancelCmd = &cobra.Command{
	Use:   "cancel <contract-id>",
	Short: "Cancel a pending dispatch contract",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newAPIClient(cmd)
		if err != nil {
			return err
		}

		path := "/api/dispatch/contracts/" + url.PathEscape(args[0]) + "/cancel"
		var result map[string]any
		if err := client.PostJSON(cmd.Context(), path, nil, &result); err != nil {
			return fmt.Errorf("cancel dispatch contract: %w", err)
		}

		fmt.Printf("Contract %s cancelled.\n", args[0])
		printDispatchJSON(result)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(dispatchCmd)
	dispatchCmd.AddCommand(dispatchCreateCmd)
	dispatchCmd.AddCommand(dispatchListCmd)
	dispatchCmd.AddCommand(dispatchShowCmd)
	dispatchCmd.AddCommand(dispatchCancelCmd)

	// dispatch create flags
	dispatchCreateCmd.Flags().String("from", "", "Source issue ID")
	dispatchCreateCmd.Flags().String("to", "", "Target issue ID (to be notified)")
	dispatchCreateCmd.Flags().String("on", "", "Comma-separated trigger event types")
	dispatchCreateCmd.Flags().String("target", "", "Optional: specific issue the contract serves")

	// dispatch list flags
	dispatchListCmd.Flags().String("status", "", "Filter by status (pending, triggered, fulfilled, cancelled)")

	_ = dispatchCreateCmd.MarkFlagRequired("from")
	_ = dispatchCreateCmd.MarkFlagRequired("to")
	_ = dispatchCreateCmd.MarkFlagRequired("on")
}

func printDispatchJSON(v any) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(v)
}
