package main

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"

	"github.com/multica-ai/multica/server/internal/cli"
)

// notificationCmd is the parent command for notification queries.
var notificationCmd = &cobra.Command{
	Use:     "notification",
	Short:   "Query pending notifications and acknowledge them",
	GroupID: groupCore,
}

// ── notification pending ──────────────────────────────────────────────────

var notificationPendingFlags struct {
	TargetType string
	TargetID   string
}

var notificationPendingCmd = &cobra.Command{
	Use:   "pending",
	Short: "List pending notifications for a target (agent or squad)",
	Args:  cobra.NoArgs,
	RunE:  runNotificationPending,
}

func runNotificationPending(cmd *cobra.Command, _ []string) error {
	client, err := newAPIClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := cli.APIContext(context.Background())
	defer cancel()

	path := fmt.Sprintf("/api/notifications/pending?target_type=%s&target_id=%s",
		notificationPendingFlags.TargetType, notificationPendingFlags.TargetID)

	var result map[string]any
	if err := client.GetJSON(ctx, path, &result); err != nil {
		return fmt.Errorf("list notifications: %w", err)
	}

	notifications, _ := result["notifications"].([]any)

	output, _ := cmd.Flags().GetString("output")
	if output == "json" {
		return cli.PrintJSON(os.Stdout, notifications)
	}

	if len(notifications) == 0 {
		fmt.Fprintln(os.Stderr, "No pending notifications.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTYPE\tSOURCE ISSUE\tMESSAGE\tCREATED")
	for _, n := range notifications {
		nm, _ := n.(map[string]any)
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			strVal(nm, "id"),
			strVal(nm, "type"),
			strVal(nm, "source_issue_id"),
			truncateMsg(strVal(nm, "message"), 50),
			strVal(nm, "created_at"),
		)
	}
	return w.Flush()
}

// ── notification acknowledge ──────────────────────────────────────────────

var notificationAckCmd = &cobra.Command{
	Use:   "acknowledge <notification-id>",
	Short: "Acknowledge a notification",
	Args:  exactArgs(1),
	RunE:  runNotificationAcknowledge,
}

func runNotificationAcknowledge(cmd *cobra.Command, args []string) error {
	client, err := newAPIClient(cmd)
	if err != nil {
		return err
	}
	ctx, cancel := cli.APIContext(context.Background())
	defer cancel()

	path := fmt.Sprintf("/api/notifications/%s/acknowledge", args[0])
	var result map[string]any
	if err := client.PostJSON(ctx, path, nil, &result); err != nil {
		return fmt.Errorf("acknowledge notification: %w", err)
	}

	output, _ := cmd.Flags().GetString("output")
	if output == "json" {
		return cli.PrintJSON(os.Stdout, result)
	}

	fmt.Fprintf(os.Stderr, "Notification %s acknowledged.\n", args[0])
	return nil
}

func truncateMsg(msg string, maxLen int) string {
	if len(msg) <= maxLen {
		return msg
	}
	return msg[:maxLen-3] + "..."
}

func init() {
	notificationPendingCmd.Flags().StringVar(&notificationPendingFlags.TargetType, "target-type", "agent", "Target type (agent|squad|member)")
	notificationPendingCmd.Flags().StringVar(&notificationPendingFlags.TargetID, "target-id", "", "Target ID (required)")
	notificationPendingCmd.MarkFlagRequired("target-id")

	notificationCmd.AddCommand(notificationPendingCmd)
	notificationCmd.AddCommand(notificationAckCmd)
}
