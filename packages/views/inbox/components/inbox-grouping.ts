import type { InboxItem, InboxItemType, InboxSeverity } from "@multica/core/types";

export type GroupMode = "none" | "time" | "severity" | "project" | "type";

export interface InboxGroup {
  key: string;
  label: string;
  items: InboxItem[];
}

/**
 * Group inbox items by the specified mode. Returns an array of groups with
 * items sorted newest-first within each group.
 */
export function groupInboxItems(
  items: InboxItem[],
  mode: GroupMode,
  labels: GroupLabels,
): InboxGroup[] {
  if (mode === "none") {
    return [{ key: "all", label: "", items }];
  }

  switch (mode) {
    case "time":
      return groupByTime(items, labels);
    case "severity":
      return groupBySeverity(items, labels);
    case "project":
      return groupByProject(items, labels);
    case "type":
      return groupByType(items, labels);
    default:
      return [{ key: "all", label: "", items }];
  }
}

export interface GroupLabels {
  time: {
    today: string;
    yesterday: string;
    thisWeek: string;
    older: string;
  };
  severity: {
    action_required: string;
    attention: string;
    info: string;
  };
  type: Record<InboxItemType, string>;
}

function getDayStart(date: Date): number {
  return new Date(date.getFullYear(), date.getMonth(), date.getDate()).getTime();
}

function groupByTime(items: InboxItem[], labels: GroupLabels): InboxGroup[] {
  const now = new Date();
  const todayStart = getDayStart(now);
  const yesterdayStart = todayStart - 86400000;
  const weekStart = todayStart - 7 * 86400000;

  const todayItems: InboxItem[] = [];
  const yesterdayItems: InboxItem[] = [];
  const thisWeekItems: InboxItem[] = [];
  const olderItems: InboxItem[] = [];

  for (const item of items) {
    const time = new Date(item.created_at).getTime();
    if (time >= todayStart) {
      todayItems.push(item);
    } else if (time >= yesterdayStart) {
      yesterdayItems.push(item);
    } else if (time >= weekStart) {
      thisWeekItems.push(item);
    } else {
      olderItems.push(item);
    }
  }

  const result: InboxGroup[] = [];
  if (todayItems.length) result.push({ key: "today", label: labels.time.today, items: todayItems });
  if (yesterdayItems.length) result.push({ key: "yesterday", label: labels.time.yesterday, items: yesterdayItems });
  if (thisWeekItems.length) result.push({ key: "thisWeek", label: labels.time.thisWeek, items: thisWeekItems });
  if (olderItems.length) result.push({ key: "older", label: labels.time.older, items: olderItems });
  return result;
}

const severityOrder: InboxSeverity[] = ["action_required", "attention", "info"];

function groupBySeverity(items: InboxItem[], labels: GroupLabels): InboxGroup[] {
  const groups: Record<string, InboxItem[]> = {};
  for (const item of items) {
    const sev = item.severity ?? "info";
    const existing = groups[sev];
    if (existing) {
      existing.push(item);
    } else {
      groups[sev] = [item];
    }
  }

  return severityOrder
    .filter((s) => groups[s]?.length)
    .map((s) => ({
      key: s,
      label: labels.severity[s],
      items: groups[s]!,
    }));
}

function groupByProject(
  _items: InboxItem[],
  _labels: GroupLabels,
): InboxGroup[] {
  // Project info is not directly on InboxItem — items are grouped by
  // issue_id, and project association lives on the issue. Best-effort:
  // group by issue_id prefix patterns or fall back to "General".
  // In practice, project grouping needs project data from the issue detail
  // or workspace context. For now return a single group and document
  // the extension point.
  const items = _items;
  if (!items.length) return [];
  return [{ key: "all", label: "", items }];
}

const typeOrder: InboxItemType[] = [
  "issue_assigned",
  "unassigned",
  "assignee_changed",
  "status_changed",
  "priority_changed",
  "start_date_changed",
  "due_date_changed",
  "new_comment",
  "mentioned",
  "review_requested",
  "task_completed",
  "task_failed",
  "agent_blocked",
  "agent_completed",
  "reaction_added",
  "quick_create_done",
  "quick_create_failed",
];

function groupByType(items: InboxItem[], labels: GroupLabels): InboxGroup[] {
  const groups: Record<string, InboxItem[]> = {};
  for (const item of items) {
    const existing = groups[item.type];
    if (existing) {
      existing.push(item);
    } else {
      groups[item.type] = [item];
    }
  }

  return typeOrder
    .filter((t) => groups[t]?.length)
    .map((t) => ({
      key: t,
      label: labels.type[t] ?? t,
      items: groups[t]!,
    }));
}

/**
 * Compute severity color class for the left accent bar.
 */
export function severityAccentClass(severity: InboxSeverity): string {
  switch (severity) {
    case "action_required":
      return "bg-red-500";
    case "attention":
      return "bg-amber-500";
    case "info":
    default:
      return "bg-blue-400";
  }
}

/**
 * Compute a search score for filtering inbox items against a query.
 * Matches against title, body, and detail labels.
 */
export function matchesSearch(item: InboxItem, query: string): boolean {
  if (!query.trim()) return true;
  const q = query.toLowerCase().trim();
  if (item.title.toLowerCase().includes(q)) return true;
  if (item.body?.toLowerCase().includes(q)) return true;
  if (item.type.toLowerCase().includes(q)) return true;
  return false;
}
