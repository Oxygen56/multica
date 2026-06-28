package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/multica-ai/multica/server/internal/cli"
)

// dispatchCmd is the parent command for dispatch contract operations.
var dispatchCmd = &cobra.Command{
	Use:     "dispatch",
	Short:   "Manage cross-squad dispatch contracts",
	GroupID: groupCore,
}

// ── dispatch create ───────────────────────────────────────────────────────

var dispatchCreateFlags struct {
	ToSquad        string
	FromIssue      string
	Title          string
	Description    string
	TargetIssue    string
	NotifyAssignee bool
	HandoffNote    string
}

var dispatchCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a cross-squad dispatch contract with callback tracking",
	Args:  cobra.NoArgs,
	RunE:  runDispatchCreate,
}

func runDispatchCreate(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := cli.APIContext(context.Background())
	defer cancel()

	body := map[string]any{
		"title":           dispatchCreateFlags.Title,
		"to_squad_id":     dispatchCreateFlags.ToSquad,
		"from_issue_id":   dispatchCreateFlags.FromIssue,
		"notify_assignee": dispatchCreateFlags.NotifyAssignee,
	}
	if dispatchCreateFlags.Description != "" {
		body["description"] = dispatchCreateFlags.Description
	}
	if dispatchCreateFlags.TargetIssue != "" {
		body["target_issue_id"] = dispatchCreateFlags.TargetIssue
	}
	if dispatchCreateFlags.HandoffNote != "" {
		body["handoff_note"] = dispatchCreateFlags.HandoffNote
	}

	var result map[string]any
	if err := client.PostJSON(ctx, "/api/dispatch", body, &result); err != nil {
		return fmt.Errorf("create dispatch: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	if output == "json" {
		return cli.PrintJSON(os.Stdout, result)
	}

	contractID, _ := result["contract_id"].(string)
	status, _ := result["status"].(string)
	issue, _ := result["issue"].(map[string]any)
	issueID := ""
	if issue != nil {
		issueID, _ = issue["id"].(string)
	}

	fmt.Fprintf(os.Stderr, "Dispatch created.\n")
	fmt.Fprintf(os.Stderr, "  Contract: %s\n", contractID)
	fmt.Fprintf(os.Stderr, "  Status:   %s\n", status)
	fmt.Fprintf(os.Stderr, "  Issue:    %s\n", issueID)
	return nil
}

// ── dispatch list ─────────────────────────────────────────────────────────

var dispatchListFlags struct {
	Status string
}

var dispatchListCmd = &cobra.Command{
	Use:   "list",
	Short: "List dispatch contracts",
	Args:  cobra.NoArgs,
	RunE:  runDispatchList,
}

func runDispatchList(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := cli.APIContext(context.Background())
	defer cancel()

	status := dispatchListFlags.Status
	if status == "" {
		status = "pending"
	}

	path := fmt.Sprintf("/api/dispatch?status=%s", status)
	var result map[string]any
	if err := client.GetJSON(ctx, path, &result); err != nil {
		return fmt.Errorf("list dispatch contracts: %w", err)
	}

	contracts, _ := result["contracts"].([]any)

	output, _ := cmd.Flags().GetString("output")
	if output == "json" {
		return cli.PrintJSON(os.Stdout, contracts)
	}

	if len(contracts) == 0 {
		fmt.Fprintln(os.Stderr, "No dispatch contracts found.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "CONTRACT ID\tFROM ISSUE\tTO ISSUE\tSTATUS")
	for _, c := range contracts {
		cm, _ := c.(map[string]any)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n",
			strVal(cm, "contract_id"),
			strVal(cm, "from_issue_id"),
			strVal(cm, "to_issue_id"),
			strVal(cm, "status"),
		)
	}
	return w.Flush()
}

// ── dispatch show ─────────────────────────────────────────────────────────

var dispatchShowCmd = &cobra.Command{
	Use:   "show <contract-id>",
	Short: "Show dispatch contract details",
	Args:  exactArgs(1),
	RunE:  runDispatchShow,
}

func runDispatchShow(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := cli.APIContext(context.Background())
	defer cancel()

	path := fmt.Sprintf("/api/dispatch/%s", args[0])
	var result map[string]any
	if err := client.GetJSON(ctx, path, &result); err != nil {
		return fmt.Errorf("get dispatch contract: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	if output == "json" {
		return cli.PrintJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Contract: %s\n", strVal(result, "contract_id"))
	fmt.Fprintf(os.Stderr, "Status:   %s\n", strVal(result, "status"))
	fmt.Fprintf(os.Stderr, "From:     %s\n", strVal(result, "from_issue_id"))
	fmt.Fprintf(os.Stderr, "To:       %s\n", strVal(result, "to_issue_id"))
	fmt.Fprintf(os.Stderr, "Target:   %s\n", strVal(result, "target_issue"))
	return nil
}

// ── dispatch cancel ───────────────────────────────────────────────────────

var dispatchCancelCmd = &cobra.Command{
	Use:   "cancel <contract-id>",
	Short: "Cancel a pending dispatch contract",
	Args:  exactArgs(1),
	RunE:  runDispatchCancel,
}

func runDispatchCancel(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := cli.APIContext(context.Background())
	defer cancel()

	path := fmt.Sprintf("/api/dispatch/%s/cancel", args[0])
	var result map[string]any
	if err := client.PostJSON(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("cancel dispatch contract: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	if output == "json" {
		return cli.PrintJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Contract %s cancelled.\n", args[0])
	return nil
}

func init() {
	// dispatch create
	dispatchCreateCmd.Flags().StringVar(&dispatchCreateFlags.ToSquad, "to-squad", "", "Target squad ID (required)")
	dispatchCreateCmd.Flags().StringVar(&dispatchCreateFlags.FromIssue, "from-issue", "", "Parent issue ID (required)")
	dispatchCreateCmd.Flags().StringVar(&dispatchCreateFlags.Title, "title", "", "Child issue title (required)")
	dispatchCreateCmd.Flags().StringVar(&dispatchCreateFlags.Description, "description", "", "Child issue description")
	dispatchCreateCmd.Flags().StringVar(&dispatchCreateFlags.TargetIssue, "target-issue", "", "Callback target issue (default: from-issue)")
	dispatchCreateCmd.Flags().BoolVar(&dispatchCreateFlags.NotifyAssignee, "notify-assignee", true, "Notify target issue assignee on completion")
	dispatchCreateCmd.Flags().StringVar(&dispatchCreateFlags.HandoffNote, "handoff-note", "", "Handoff note for the dispatched squad")
	dispatchCreateCmd.MarkFlagRequired("to-squad")
	dispatchCreateCmd.MarkFlagRequired("from-issue")
	dispatchCreateCmd.MarkFlagRequired("title")

	// dispatch list
	dispatchListCmd.Flags().StringVar(&dispatchListFlags.Status, "status", "pending", "Filter by status (pending|triggered|fulfilled|cancelled)")

	dispatchCmd.AddCommand(dispatchCreateCmd)
	dispatchCmd.AddCommand(dispatchListCmd)
	dispatchCmd.AddCommand(dispatchShowCmd)
	dispatchCmd.AddCommand(dispatchCancelCmd)
}
