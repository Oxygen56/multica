"use client";

import { StatusIcon } from "../../issues/components";
import { ActorAvatar } from "../../common/actor-avatar";
import { Archive, CheckCheck, ExternalLink } from "lucide-react";
import type { InboxItem } from "@multica/core/types";
import { InboxDetailLabel } from "./inbox-detail-label";
import { getInboxDisplayTitle } from "./inbox-display";
import { severityAccentClass } from "./inbox-grouping";
import { InboxSelectCheckbox } from "./inbox-batch-bar";
import { useT } from "../../i18n";
import { cn } from "@multica/ui/lib/utils";

// Hook returning a localized relative-time formatter — the i18n equivalent
// of the previous static `timeAgo` function. Returning a function (rather
// than a string) keeps call-site usage identical: `timeAgo(dateStr)`.
export function useTimeAgo() {
  const { t } = useT("inbox");
  return (dateStr: string): string => {
    const diff = Date.now() - new Date(dateStr).getTime();
    const minutes = Math.floor(diff / 60000);
    if (minutes < 1) return t(($) => $.list.time.just_now);
    if (minutes < 60) return t(($) => $.list.time.minutes, { count: minutes });
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return t(($) => $.list.time.hours, { count: hours });
    const days = Math.floor(hours / 24);
    return t(($) => $.list.time.days, { count: days });
  };
}

export interface InboxListItemProps {
  item: InboxItem;
  isSelected: boolean;
  isFocused: boolean;
  density: "comfortable" | "compact";
  batchMode: boolean;
  batchSelected: boolean;
  onSelect: () => void;
  onArchive: () => void;
  onMarkRead: () => void;
  onToggleBatchSelect: (id: string, checked: boolean) => void;
  onOpenIssue: () => void;
}

export function InboxListItem({
  item,
  isSelected,
  isFocused,
  density,
  batchMode,
  batchSelected,
  onSelect,
  onArchive,
  onMarkRead,
  onToggleBatchSelect,
  onOpenIssue,
}: InboxListItemProps) {
  const { t } = useT("inbox");
  const timeAgo = useTimeAgo();
  const displayTitle = getInboxDisplayTitle(item);
  const accentClass = severityAccentClass(item.severity);

  return (
    <div
      className={cn(
        "group flex items-stretch transition-colors",
        isSelected && "bg-accent",
        !isSelected && "hover:bg-accent/50",
        isFocused && "ring-1 ring-inset ring-ring",
        density === "compact" ? "py-0.5" : "py-1.5",
      )}
      role="listitem"
      aria-selected={isSelected}
      aria-label={displayTitle}
    >
      {/* Severity color accent bar (left edge) */}
      <div
        className={cn(
          "w-1 shrink-0 transition-opacity",
          accentClass,
          !isSelected && !isFocused && "opacity-60 group-hover:opacity-100",
        )}
        aria-hidden="true"
      />

      {/* Batch select checkbox */}
      {batchMode && (
        <InboxSelectCheckbox
          id={item.id}
          checked={batchSelected}
          onChange={onToggleBatchSelect}
        />
      )}

      {/* Main click target */}
      <button
        type="button"
        onClick={batchMode ? () => onToggleBatchSelect(item.id, !batchSelected) : onSelect}
        className={cn(
          "flex flex-1 min-w-0 items-center gap-3 text-left",
          density === "compact" ? "px-2 py-0.5" : "px-4 py-2.5",
        )}
        tabIndex={isFocused ? 0 : -1}
        data-inbox-item-id={item.id}
      >
        {/* Actor avatar */}
        <ActorAvatar
          actorType={item.actor_type ?? item.recipient_type}
          actorId={item.actor_id ?? item.recipient_id}
          size={density === "compact" ? 24 : 28}
          enableHoverCard
        />

        {/* Primary + Secondary tiers */}
        <div className="min-w-0 flex-1">
          {/* Primary tier: title + unread indicator */}
          <div className="flex items-center justify-between gap-2">
            <div className="flex min-w-0 items-center gap-1.5">
              {!item.read && (
                <span
                  className={cn(
                    "h-1.5 w-1.5 shrink-0 rounded-full bg-brand",
                    density === "compact" && "h-1 w-1",
                  )}
                  aria-label={t(($) => $.list.unread)}
                />
              )}
              <span
                className={cn(
                  "truncate",
                  density === "compact" ? "text-xs" : "text-sm",
                  !item.read ? "font-medium" : "text-muted-foreground",
                )}
              >
                {displayTitle}
              </span>
            </div>

            {/* Tertiary actions: hover-revealed */}
            <div className="flex shrink-0 items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
              {/* Mark read */}
              {!item.read && (
                <span
                  role="button"
                  tabIndex={-1}
                  title={t(($) => $.list.mark_read_tooltip)}
                  onClick={(e) => {
                    e.stopPropagation();
                    onMarkRead();
                  }}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.stopPropagation();
                      onMarkRead();
                    }
                  }}
                  className="rounded p-0.5 text-muted-foreground hover:bg-accent hover:text-foreground"
                >
                  <CheckCheck className="h-3.5 w-3.5" />
                </span>
              )}
              {/* Archive */}
              <span
                role="button"
                tabIndex={-1}
                title={t(($) => $.list.archive_tooltip)}
                onClick={(e) => {
                  e.stopPropagation();
                  onArchive();
                }}
                onKeyDown={(e) => {
                  if (e.key === "Enter" || e.key === " ") {
                    e.stopPropagation();
                    onArchive();
                  }
                }}
                className="rounded p-0.5 text-muted-foreground hover:bg-accent hover:text-foreground"
              >
                <Archive className="h-3.5 w-3.5" />
              </span>
              {/* Open issue */}
              {item.issue_id && (
                <span
                  role="button"
                  tabIndex={-1}
                  title={t(($) => $.list.open_issue_tooltip)}
                  onClick={(e) => {
                    e.stopPropagation();
                    onOpenIssue();
                  }}
                  onKeyDown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                      e.stopPropagation();
                      onOpenIssue();
                    }
                  }}
                  className="rounded p-0.5 text-muted-foreground hover:bg-accent hover:text-foreground"
                >
                  <ExternalLink className="h-3.5 w-3.5" />
                </span>
              )}
            </div>
          </div>

          {/* Secondary tier: detail label + timestamp */}
          <div className="mt-0.5 flex items-center justify-between gap-2">
            <p
              className={cn(
                "min-w-0 overflow-hidden text-ellipsis whitespace-nowrap",
                density === "compact" ? "text-[10px]" : "text-xs",
                item.read
                  ? "text-muted-foreground/60"
                  : "text-muted-foreground",
              )}
            >
              <InboxDetailLabel item={item} />
            </p>

            {/* Status icon */}
            <div className="flex shrink-0 items-center gap-1">
              {item.issue_status && (
                <StatusIcon
                  status={item.issue_status}
                  className={cn(
                    "shrink-0",
                    density === "compact" ? "h-3 w-3" : "h-3.5 w-3.5",
                  )}
                />
              )}
              <span
                className={cn(
                  "shrink-0 tabular-nums",
                  density === "compact" ? "text-[10px]" : "text-xs",
                  item.read
                    ? "text-muted-foreground/60"
                    : "text-muted-foreground",
                )}
              >
                {timeAgo(item.created_at)}
              </span>
            </div>
          </div>
        </div>
      </button>
    </div>
  );
}
