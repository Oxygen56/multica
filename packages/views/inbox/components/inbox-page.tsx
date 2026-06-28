"use client";

import { useState, useEffect, useCallback, useMemo, useRef } from "react";
import { useDefaultLayout } from "react-resizable-panels";
import { useQuery } from "@tanstack/react-query";
import { useWorkspaceId } from "@multica/core/hooks";
import { useWorkspacePaths } from "@multica/core/paths";
import { useModalStore } from "@multica/core/modals";
import { useIssueDraftStore } from "@multica/core/issues/stores/draft-store";
import {
  inboxListOptions,
  deduplicateInboxItems,
  useInboxUnreadCount,
} from "@multica/core/inbox/queries";
import {
  useMarkInboxRead,
  useArchiveInbox,
  useMarkAllInboxRead,
  useArchiveAllInbox,
  useArchiveAllReadInbox,
  useArchiveCompletedInbox,
} from "@multica/core/inbox/mutations";

import { IssueDetail } from "../../issues/components";
import { ErrorBoundary } from "@multica/ui/components/common/error-boundary";
import { useNavigation } from "../../navigation";
import { toast } from "sonner";
import {
  MoreHorizontal,
  Inbox,
  CheckCheck,
  Archive,
  BookCheck,
  ListChecks,
  ArrowLeft,
} from "lucide-react";
import type { InboxItem } from "@multica/core/types";
import { Button } from "@multica/ui/components/ui/button";
import {
  ResizablePanelGroup,
  ResizablePanel,
  ResizableHandle,
} from "@multica/ui/components/ui/resizable";
import { Skeleton } from "@multica/ui/components/ui/skeleton";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
} from "@multica/ui/components/ui/dropdown-menu";
import { useIsMobile } from "@multica/ui/hooks/use-mobile";
import { PageHeader } from "../../layout/page-header";
import { InboxListItem, useTimeAgo } from "./inbox-list-item";
import { useTypeLabels } from "./inbox-detail-label";
import { getInboxDisplayTitle } from "./inbox-display";
import { useT } from "../../i18n";
import {
  InboxToolbar,
  type InboxToolbarState,
} from "./inbox-toolbar";
import {
  groupInboxItems,
  matchesSearch,
  type GroupLabels,
} from "./inbox-grouping";
import { InboxBatchBar } from "./inbox-batch-bar";
import { useInboxKeyboardNav } from "./inbox-keyboard-nav";

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const DEFAULT_TOOLBAR_STATE: InboxToolbarState = {
  searchQuery: "",
  groupMode: "none",
  unreadOnly: false,
  density: "comfortable",
};

// ---------------------------------------------------------------------------
// InboxPage
// ---------------------------------------------------------------------------

export function InboxPage() {
  const { t } = useT("inbox");
  const { searchParams, replace } = useNavigation();
  const urlIssue = searchParams.get("issue") ?? "";
  const wsPaths = useWorkspacePaths();

  const [selectedKey, setSelectedKeyState] = useState(() => urlIssue);

  // Sync from URL when searchParams change (e.g. navigation)
  useEffect(() => {
    setSelectedKeyState(urlIssue);
  }, [urlIssue]);

  const wsId = useWorkspaceId();
  const { data: rawItems = [], isLoading: loading } = useQuery(inboxListOptions(wsId));
  const items = useMemo(() => deduplicateInboxItems(rawItems), [rawItems]);

  const selected = items.find((i) => (i.issue_id ?? i.id) === selectedKey) ?? null;

  // Track the last key we actually resolved against the inbox list. Lets the
  // fallback effect distinguish "shared-link to a notification not in our
  // inbox" (never resolved → redirect to the issue page) from "item was in
  // our inbox and just got removed" (was resolved → stay on /inbox).
  const lastResolvedKeyRef = useRef<string>("");
  useEffect(() => {
    if (selected) lastResolvedKeyRef.current = selectedKey;
  }, [selected, selectedKey]);

  const setSelectedKey = useCallback((key: string) => {
    setSelectedKeyState(key);
    const inboxPath = wsPaths.inbox();
    const url = key ? `${inboxPath}?issue=${key}` : inboxPath;
    replace(url);
  }, [replace, wsPaths]);

  // Shared inbox links (?issue=<id>) may point to notifications not in this
  // user's inbox (archived, or never received). Fall back to the issue page
  // so the URL still resolves to something meaningful. But if the key was
  // previously resolvable (e.g. the issue was just deleted in another tab
  // and `onInboxIssueDeleted` pruned the cache), the issue detail would 404
  // too — clear the selection and stay on /inbox instead.
  useEffect(() => {
    if (loading) return;
    if (!selectedKey) return;
    if (selected) return;
    if (lastResolvedKeyRef.current === selectedKey) {
      setSelectedKey("");
      return;
    }
    replace(wsPaths.issueDetail(selectedKey));
  }, [loading, selectedKey, selected, replace, wsPaths, setSelectedKey]);

  const { defaultLayout, onLayoutChanged } = useDefaultLayout({
    id: "multica_inbox_layout",
  });

  const isMobile = useIsMobile();
  const unreadCount = useInboxUnreadCount(wsId);

  const markReadMutation = useMarkInboxRead();
  const archiveMutation = useArchiveInbox();
  const markAllReadMutation = useMarkAllInboxRead();
  const archiveAllMutation = useArchiveAllInbox();
  const archiveAllReadMutation = useArchiveAllReadInbox();
  const archiveCompletedMutation = useArchiveCompletedInbox();
  const timeAgo = useTimeAgo();
  const typeLabels = useTypeLabels();

  // -- Toolbar state ---------------------------------------------------------
  const [toolbar, setToolbar] = useState<InboxToolbarState>(DEFAULT_TOOLBAR_STATE);
  const searchInputRef = useRef<HTMLInputElement>(null);

  // -- Batch selection state -------------------------------------------------
  const [batchMode, setBatchMode] = useState(false);
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  // -- Keyboard navigation state ---------------------------------------------
  const [focusedIndex, setFocusedIndex] = useState(-1);
  const listContainerRef = useRef<HTMLDivElement>(null);

  // -- Filtering & grouping --------------------------------------------------
  const groupLabels: GroupLabels = useMemo(
    () => ({
      time: {
        today: t(($) => $.groups.today),
        yesterday: t(($) => $.groups.yesterday),
        thisWeek: t(($) => $.groups.this_week),
        older: t(($) => $.groups.older),
      },
      severity: {
        action_required: t(($) => $.groups.action_required),
        attention: t(($) => $.groups.attention),
        info: t(($) => $.groups.info),
      },
      type: typeLabels,
    }),
    [t, typeLabels],
  );

  const filteredItems = useMemo(() => {
    let result = items;
    if (toolbar.unreadOnly) {
      result = result.filter((i) => !i.read);
    }
    if (toolbar.searchQuery.trim()) {
      result = result.filter((i) => matchesSearch(i, toolbar.searchQuery));
    }
    return result;
  }, [items, toolbar.unreadOnly, toolbar.searchQuery]);

  const groups = useMemo(
    () => groupInboxItems(filteredItems, toolbar.groupMode, groupLabels),
    [filteredItems, toolbar.groupMode, groupLabels],
  );

  // Flattened for keyboard nav — items in the order they're rendered
  const flatRenderedItems = useMemo(
    () => groups.flatMap((g) => g.items),
    [groups],
  );

  // Reset focusedIndex when the list changes
  useEffect(() => {
    if (focusedIndex >= flatRenderedItems.length) {
      setFocusedIndex(Math.max(0, flatRenderedItems.length - 1));
    }
  }, [flatRenderedItems.length, focusedIndex]);

  // Auto-mark-read whenever a selected item is unread — covers both click-
  // to-select and URL-param-select (e.g. OS notification click on desktop).
  // The mutation flips `read: true` optimistically, so this effect settles
  // in one pass and can't loop. Kept in a `useEffect` rather than inlined
  // in handleSelect so URL-driven selection triggers it too.
  const markReadMutate = markReadMutation.mutate;
  const selectedId = selected?.id;
  const selectedRead = selected?.read;
  useEffect(() => {
    if (!selectedId || selectedRead) return;
    markReadMutate(selectedId, {
      onError: (err) =>
        toast.error(
          err instanceof Error && err.message
            ? err.message
            : t(($) => $.errors.mark_read_failed),
        ),
    });
  }, [selectedId, selectedRead, markReadMutate, t]);

  // -- Handlers ---------------------------------------------------------------

  const handleSelect = (item: InboxItem) => {
    setSelectedKey(item.issue_id ?? item.id);
    setFocusedIndex(flatRenderedItems.indexOf(item));
  };

  const handleArchive = useCallback(
    (id: string) => {
      const idx = items.findIndex((i) => i.id === id);
      const archived = idx >= 0 ? items[idx] : null;
      const wasSelected =
        !!archived && (archived.issue_id ?? archived.id) === selectedKey;
      if (wasSelected) {
        // List is sorted newest-first; prefer the next (older) item, fall back
        // to the previous (newer) one when archiving at the bottom, and only
        // clear the selection when nothing else is left.
        const next = items[idx + 1] ?? items[idx - 1] ?? null;
        setSelectedKey(next ? (next.issue_id ?? next.id) : "");
      }
      archiveMutation.mutate(id, {
        onError: (err) =>
          toast.error(
            err instanceof Error && err.message
              ? err.message
              : t(($) => $.errors.archive_failed),
          ),
      });
    },
    [items, selectedKey, archiveMutation, setSelectedKey, t],
  );

  const handleMarkRead = useCallback(
    (id: string) => {
      markReadMutation.mutate(id, {
        onError: (err) =>
          toast.error(
            err instanceof Error && err.message
              ? err.message
              : t(($) => $.errors.mark_read_failed),
          ),
      });
    },
    [markReadMutation, t],
  );

  const handleOpenIssue = useCallback(
    (issueId: string | null) => {
      if (!issueId) return;
      replace(wsPaths.issueDetail(issueId));
    },
    [replace, wsPaths],
  );

  // Batch operations
  const handleMarkAllRead = () => {
    markAllReadMutation.mutate(undefined, {
      onError: (err) =>
        toast.error(
          err instanceof Error && err.message
            ? err.message
            : t(($) => $.errors.mark_all_read_failed),
        ),
    });
  };

  const handleArchiveAll = () => {
    setSelectedKey("");
    archiveAllMutation.mutate(undefined, {
      onError: (err) =>
        toast.error(
          err instanceof Error && err.message
            ? err.message
            : t(($) => $.errors.archive_all_failed),
        ),
    });
  };

  const handleArchiveAllRead = () => {
    const readKeys = items.filter((i) => i.read).map((i) => i.issue_id ?? i.id);
    if (readKeys.includes(selectedKey)) setSelectedKey("");
    archiveAllReadMutation.mutate(undefined, {
      onError: (err) =>
        toast.error(
          err instanceof Error && err.message
            ? err.message
            : t(($) => $.errors.archive_all_read_failed),
        ),
    });
  };

  const handleArchiveCompleted = () => {
    setSelectedKey("");
    archiveCompletedMutation.mutate(undefined, {
      onError: (err) =>
        toast.error(
          err instanceof Error && err.message
            ? err.message
            : t(($) => $.errors.archive_completed_failed),
        ),
    });
  };

  // Batch selection handlers
  const handleToggleBatchSelect = useCallback((id: string, checked: boolean) => {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (checked) {
        next.add(id);
      } else {
        next.delete(id);
      }
      return next;
    });
  }, []);

  const handleToggleSelectAll = useCallback(() => {
    const renderedIds = flatRenderedItems.map((i) => i.id);
    if (selectedIds.size === renderedIds.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(renderedIds));
    }
  }, [flatRenderedItems, selectedIds.size]);

  const handleMarkReadSelected = useCallback(() => {
    for (const id of selectedIds) {
      markReadMutation.mutate(id);
    }
    setSelectedIds(new Set());
    setBatchMode(false);
  }, [selectedIds, markReadMutation]);

  const handleArchiveSelected = useCallback(() => {
    for (const id of selectedIds) {
      archiveMutation.mutate(id);
    }
    // Clear selection if selected item was archived
    const selectedInBatch = flatRenderedItems.find(
      (i) => selectedIds.has(i.id) && (i.issue_id ?? i.id) === selectedKey,
    );
    if (selectedInBatch) setSelectedKey("");
    setSelectedIds(new Set());
    setBatchMode(false);
  }, [selectedIds, archiveMutation, flatRenderedItems, selectedKey, setSelectedKey]);

  const handleClearSelection = useCallback(() => {
    setSelectedIds(new Set());
    setBatchMode(false);
  }, []);

  // Keyboard navigation actions
  const keyboardActions = useMemo(
    () => ({
      setFocusedIndex,
      moveFocusUp: () =>
        setFocusedIndex((i) => (i > 0 ? i - 1 : i)),
      moveFocusDown: () =>
        setFocusedIndex((i) =>
          i < flatRenderedItems.length - 1 ? i + 1 : i,
        ),
      archiveFocused: () => {
        const item = flatRenderedItems[focusedIndex];
        if (item) handleArchive(item.id);
      },
      focusSearch: () => {
        searchInputRef.current?.focus();
      },
      toggleMultiSelect: () => {
        setBatchMode((m) => !m);
        if (batchMode) setSelectedIds(new Set());
      },
      toggleSelectFocused: () => {
        const item = flatRenderedItems[focusedIndex];
        if (item) {
          setSelectedIds((prev) => {
            const next = new Set(prev);
            if (next.has(item.id)) {
              next.delete(item.id);
            } else {
              next.add(item.id);
            }
            return next;
          });
        }
      },
    }),
    [flatRenderedItems, focusedIndex, handleArchive, batchMode],
  );

  const keyboardNavState = useMemo(
    () => ({
      focusedIndex,
      multiSelectMode: batchMode,
      selectedIds,
    }),
    [focusedIndex, batchMode, selectedIds],
  );

  useInboxKeyboardNav(
    flatRenderedItems,
    keyboardNavState,
    keyboardActions,
    !isMobile,
  );

  // Scroll focused item into view
  useEffect(() => {
    if (focusedIndex < 0 || !listContainerRef.current) return;
    const focusedEl = listContainerRef.current.querySelector(
      `[data-inbox-item-id="${flatRenderedItems[focusedIndex]?.id}"]`,
    );
    focusedEl?.scrollIntoView({ block: "nearest" });
  }, [focusedIndex, flatRenderedItems]);

  // -- Sub-components ---------------------------------------------------------

  const toolbarEl = (
    <InboxToolbar
      state={toolbar}
      unreadCount={unreadCount}
      totalCount={filteredItems.length}
      onChange={(update) => setToolbar((prev) => ({ ...prev, ...update }))}
      onClearSearch={() => setToolbar((prev) => ({ ...prev, searchQuery: "" }))}
    />
  );

  const batchBar = (
    <InboxBatchBar
      selectedIds={selectedIds}
      totalCount={flatRenderedItems.length}
      onToggleSelectAll={handleToggleSelectAll}
      onMarkReadSelected={handleMarkReadSelected}
      onArchiveSelected={handleArchiveSelected}
      onClearSelection={handleClearSelection}
    />
  );

  const listHeader = (
    <PageHeader className="justify-between">
      <div className="flex items-center gap-2">
        <h1 className="text-sm font-semibold">{t(($) => $.page.title)}</h1>
        {unreadCount > 0 && (
          <span className="text-xs text-muted-foreground tabular-nums">
            {unreadCount}
          </span>
        )}
      </div>
      <div className="flex items-center gap-1">
        {/* Batch mode toggle */}
        <Button
          variant="ghost"
          size="sm"
          className="h-7 text-xs text-muted-foreground"
          onClick={() => {
            setBatchMode((m) => !m);
            if (batchMode) setSelectedIds(new Set());
          }}
          aria-pressed={batchMode}
          aria-label={t(($) => $.batch.toggle_label)}
        >
          {batchMode
            ? t(($) => $.batch.done)
            : t(($) => $.batch.select)}
        </Button>

        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button
                variant="ghost"
                size="icon-sm"
                className="text-muted-foreground"
              />
            }
          >
            <MoreHorizontal className="h-4 w-4" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-auto">
            <DropdownMenuItem onClick={handleMarkAllRead}>
              <CheckCheck className="h-4 w-4" />
              {t(($) => $.menu.mark_all_read)}
            </DropdownMenuItem>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={handleArchiveAll}>
              <Archive className="h-4 w-4" />
              {t(($) => $.menu.archive_all)}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleArchiveAllRead}>
              <BookCheck className="h-4 w-4" />
              {t(($) => $.menu.archive_all_read)}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={handleArchiveCompleted}>
              <ListChecks className="h-4 w-4" />
              {t(($) => $.menu.archive_completed)}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </PageHeader>
  );

  const listBody = items.length === 0 ? (
    <div
      className="flex flex-col items-center justify-center py-16 text-muted-foreground"
      role="status"
    >
      <Inbox className="mb-3 h-8 w-8 text-muted-foreground/50" aria-hidden="true" />
      <p className="text-sm">{t(($) => $.list.empty)}</p>
    </div>
  ) : groups.length === 0 ? (
    <div
      className="flex flex-col items-center justify-center py-16 text-muted-foreground"
      role="status"
    >
      <p className="text-sm">{t(($) => $.list.no_matches)}</p>
    </div>
  ) : (
    <div ref={listContainerRef} role="list" aria-label={t(($) => $.list.aria_label)}>
      {groups.map((group) => (
        <div key={group.key}>
          {/* Group header */}
          {group.label && group.items.length > 0 && (
            <div
              className="px-4 py-2 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground/70 select-none"
              role="presentation"
            >
              {group.label}
              <span className="ml-1 font-normal normal-case text-[10px]">
                ({group.items.length})
              </span>
            </div>
          )}
          {group.items.map((item) => (
            <InboxListItem
              key={item.id}
              item={item}
              isSelected={(item.issue_id ?? item.id) === selectedKey}
              isFocused={
                focusedIndex >= 0 &&
                flatRenderedItems[focusedIndex]?.id === item.id
              }
              density={toolbar.density}
              batchMode={batchMode}
              batchSelected={selectedIds.has(item.id)}
              onSelect={() => handleSelect(item)}
              onArchive={() => handleArchive(item.id)}
              onMarkRead={() => handleMarkRead(item.id)}
              onToggleBatchSelect={handleToggleBatchSelect}
              onOpenIssue={() => handleOpenIssue(item.issue_id)}
            />
          ))}
        </div>
      ))}
    </div>
  );

  const detailContent = selected?.issue_id ? (
    // Key by issue_id (not inbox-item id): a new comment/reaction generates a
    // new inbox notification for the same issue, and the dedup helper picks the
    // newest one — keying on its id would remount IssueDetail on every event,
    // wiping the comment composer draft and resetting scroll position.
    <ErrorBoundary resetKeys={[selected.issue_id]}>
      <IssueDetail
        key={selected.issue_id}
        issueId={selected.issue_id}
        defaultSidebarOpen={false}
        layoutId="multica_inbox_issue_detail_layout"
        highlightCommentId={selected.details?.comment_id ?? undefined}
        onDelete={() => {
          setSelectedKey("");
        }}
        onDone={() => {
          handleArchive(selected.id);
        }}
      />
    </ErrorBoundary>
  ) : selected ? (
    <div className="p-6" role="article" aria-label={getInboxDisplayTitle(selected)}>
      <h2 className="text-lg font-semibold">{getInboxDisplayTitle(selected)}</h2>
      <p className="mt-1 text-sm text-muted-foreground">
        {typeLabels[selected.type]} · {timeAgo(selected.created_at)}
      </p>
      {selected.body && (
        <div className="mt-4 whitespace-pre-wrap text-sm leading-relaxed text-foreground/80">
          {selected.body}
        </div>
      )}
      {selected.type === "quick_create_failed" && selected.details?.original_prompt && (
        <div className="mt-4 rounded-md border bg-muted/40 p-3">
          <p className="text-xs font-medium text-muted-foreground">
            {t(($) => $.detail.original_input)}
          </p>
          <p className="mt-1 whitespace-pre-wrap text-sm">{selected.details.original_prompt}</p>
        </div>
      )}
      <div className="mt-4 flex gap-2">
        {selected.type === "quick_create_failed" && (
          <Button
            size="sm"
            onClick={() => {
              const prompt = selected.details?.original_prompt ?? "";
              const agentId = selected.details?.agent_id;
              useIssueDraftStore.getState().setDraft({
                description: prompt,
                ...(agentId
                  ? { assigneeType: "agent" as const, assigneeId: agentId }
                  : {}),
              });
              useModalStore.getState().open("create-issue");
            }}
          >
            {t(($) => $.detail.edit_advanced)}
          </Button>
        )}
        <Button
          variant="outline"
          size="sm"
          onClick={() => handleArchive(selected.id)}
        >
          <Archive className="mr-1.5 h-3.5 w-3.5" />
          {t(($) => $.detail.archive)}
        </Button>
      </div>
    </div>
  ) : null;

  // -- Mobile layout: list / detail toggle -----------------------------------

  if (isMobile) {
    if (loading) {
      return (
        <div className="flex flex-1 flex-col min-h-0">
          <div className="flex h-12 shrink-0 items-center border-b px-4">
            <Skeleton className="h-5 w-16" />
          </div>
          <div className="flex-1 min-h-0 overflow-y-auto space-y-1 p-2">
            {Array.from({ length: 5 }).map((_, i) => (
              <div key={i} className="flex items-center gap-3 px-4 py-2.5">
                <Skeleton className="h-7 w-7 shrink-0 rounded-full" />
                <div className="flex-1 space-y-2">
                  <Skeleton className="h-4 w-3/4" />
                  <Skeleton className="h-3 w-1/2" />
                </div>
              </div>
            ))}
          </div>
        </div>
      );
    }

    // Mobile: show detail full-screen when an item is selected
    if (selected) {
      return (
        <div className="flex flex-1 flex-col min-h-0">
          <div className="flex h-12 shrink-0 items-center border-b px-2">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setSelectedKey("")}
              className="gap-1.5 text-muted-foreground"
              aria-label={t(($) => $.page.back)}
            >
              <ArrowLeft className="h-4 w-4" aria-hidden="true" />
              {t(($) => $.page.back)}
            </Button>
          </div>
          <div className="flex-1 min-h-0 overflow-y-auto">
            {detailContent}
          </div>
        </div>
      );
    }

    // Mobile: full-screen list with toolbar
    return (
      <div className="flex flex-1 flex-col min-h-0">
        {listHeader}
        {toolbarEl}
        <div className="flex-1 min-h-0 overflow-y-auto">
          {listBody}
        </div>
        {batchBar}
      </div>
    );
  }

  // -- Desktop layout: resizable two-panel -----------------------------------

  if (loading) {
    return (
      <ResizablePanelGroup orientation="horizontal" className="flex-1 min-h-0" defaultLayout={defaultLayout} onLayoutChanged={onLayoutChanged}>
        <ResizablePanel id="list" defaultSize={320} minSize={240} maxSize={480} groupResizeBehavior="preserve-pixel-size">
          <div className="flex flex-col border-r h-full">
            <div className="flex h-12 shrink-0 items-center border-b px-4">
              <Skeleton className="h-5 w-16" />
            </div>
            <div className="flex-1 min-h-0 overflow-y-auto space-y-1 p-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <div key={i} className="flex items-center gap-3 px-4 py-2.5">
                  <Skeleton className="h-7 w-7 shrink-0 rounded-full" />
                  <div className="flex-1 space-y-2">
                    <Skeleton className="h-4 w-3/4" />
                    <Skeleton className="h-3 w-1/2" />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </ResizablePanel>
        <ResizableHandle />
        <ResizablePanel id="detail" minSize="40%">
          <div className="p-6">
            <Skeleton className="h-6 w-48" />
            <Skeleton className="mt-4 h-4 w-32" />
          </div>
        </ResizablePanel>
      </ResizablePanelGroup>
    );
  }

  return (
    <ResizablePanelGroup
      orientation="horizontal"
      className="flex-1 min-h-0"
      defaultLayout={defaultLayout}
      onLayoutChanged={onLayoutChanged}
    >
      <ResizablePanel
        id="list"
        defaultSize={320}
        minSize={240}
        maxSize={480}
        groupResizeBehavior="preserve-pixel-size"
      >
        <div className="flex flex-col border-r h-full">
          {listHeader}
          {toolbarEl}
          <div className="flex-1 min-h-0 overflow-y-auto">
            {listBody}
          </div>
          {batchBar}
        </div>
      </ResizablePanel>
      <ResizableHandle />
      <ResizablePanel id="detail" minSize="40%">
        <div className="flex flex-col min-h-0 h-full">
          {detailContent ?? (
            <div
              className="flex h-full flex-col items-center justify-center text-muted-foreground"
              role="status"
            >
              <Inbox className="mb-3 h-10 w-10 text-muted-foreground/30" aria-hidden="true" />
              <p className="text-sm">
                {items.length === 0
                  ? t(($) => $.detail.empty)
                  : t(($) => $.detail.select_prompt)}
              </p>
            </div>
          )}
        </div>
      </ResizablePanel>
    </ResizablePanelGroup>
  );
}
