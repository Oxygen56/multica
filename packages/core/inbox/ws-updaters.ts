import type { QueryClient } from "@tanstack/react-query";
import { inboxKeys } from "./queries";
import type { InboxItem, IssueStatus } from "../types";

/**
 * Handle a new inbox item arriving via WebSocket.
 * Inserts the new item at the top of the cache without a full refetch
 * for instant feedback. Falls back to invalidateQueries so all observers
 * stay consistent.
 */
export function onInboxNew(
  qc: QueryClient,
  wsId: string,
  item: InboxItem,
) {
  const key = inboxKeys.list(wsId);

  // Insert optimistically at position 0 within the cached list
  qc.setQueryData<InboxItem[]>(key, (old) => {
    if (!old) return [item];
    // Dedup: if an item with the same id already exists, replace it
    const filtered = old.filter((i) => i.id !== item.id);
    // Newest first: insert at top
    return [item, ...filtered].sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
  });

  // Then invalidate to ensure server-truth alignment
  qc.invalidateQueries({ queryKey: key });
}

/**
 * Handle a batch of new inbox items arriving via WebSocket.
 * Inserts them at the top, deduplicating by id.
 */
export function onInboxBatch(
  qc: QueryClient,
  wsId: string,
  newItems: InboxItem[],
) {
  const key = inboxKeys.list(wsId);

  qc.setQueryData<InboxItem[]>(key, (old) => {
    if (!old) return newItems;
    const existingIds = new Set(newItems.map((i) => i.id));
    const filtered = old.filter((i) => !existingIds.has(i.id));
    return [...newItems, ...filtered].sort(
      (a, b) =>
        new Date(b.created_at).getTime() - new Date(a.created_at).getTime(),
    );
  });

  qc.invalidateQueries({ queryKey: key });
}

export function onInboxIssueStatusChanged(
  qc: QueryClient,
  wsId: string,
  issueId: string,
  status: IssueStatus,
) {
  qc.setQueryData<InboxItem[]>(inboxKeys.list(wsId), (old) =>
    old?.map((i) =>
      i.issue_id === issueId ? { ...i, issue_status: status } : i,
    ),
  );
}

// Mirrors the DB-level ON DELETE CASCADE on inbox_item.issue_id: when an issue
// is deleted, all inbox items that referenced it are gone server-side, so drop
// them from the cache too.
export function onInboxIssueDeleted(
  qc: QueryClient,
  wsId: string,
  issueId: string,
) {
  qc.setQueryData<InboxItem[]>(inboxKeys.list(wsId), (old) =>
    old?.filter((i) => i.issue_id !== issueId),
  );
}

/**
 * Restore the last viewed state after WebSocket reconnection.
 * Invalidates queries to refresh from server to catch any missed events.
 */
export function onInboxReconnected(qc: QueryClient, wsId: string) {
  qc.invalidateQueries({ queryKey: inboxKeys.list(wsId) });
}

export function onInboxInvalidate(qc: QueryClient, wsId: string) {
  qc.invalidateQueries({ queryKey: inboxKeys.list(wsId) });
}
