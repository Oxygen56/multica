"use client";

import { Search, LayoutGrid, Filter, ListFilter, X } from "lucide-react";
import { Button } from "@multica/ui/components/ui/button";
import { Input } from "@multica/ui/components/ui/input";
import {
  DropdownMenu,
  DropdownMenuTrigger,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuCheckboxItem,
} from "@multica/ui/components/ui/dropdown-menu";
import { Toggle } from "@multica/ui/components/ui/toggle";
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from "@multica/ui/components/ui/tooltip";
import { useT } from "../../i18n";

export type GroupMode = "none" | "time" | "severity" | "project" | "type";
export type DensityMode = "comfortable" | "compact";

export interface InboxToolbarState {
  searchQuery: string;
  groupMode: GroupMode;
  unreadOnly: boolean;
  density: DensityMode;
}

export function InboxToolbar({
  state,
  unreadCount,
  totalCount,
  onChange,
  onClearSearch,
}: {
  state: InboxToolbarState;
  unreadCount: number;
  totalCount: number;
  onChange: (update: Partial<InboxToolbarState>) => void;
  onClearSearch: () => void;
}) {
  const { t } = useT("inbox");

  const groupLabels: Record<GroupMode, string> = {
    none: t(($) => $.toolbar.group_none),
    time: t(($) => $.toolbar.group_time),
    severity: t(($) => $.toolbar.group_severity),
    project: t(($) => $.toolbar.group_project),
    type: t(($) => $.toolbar.group_type),
  };

  return (
    <div
      className="flex items-center gap-1.5 px-3 py-2 border-b shrink-0"
      role="toolbar"
      aria-label={t(($) => $.toolbar.aria_label)}
    >
      {/* Search */}
      <div className="relative flex-1 min-w-0 max-w-[260px]">
        <Search className="absolute left-2 top-1/2 -translate-y-1/2 h-3.5 w-3.5 text-muted-foreground pointer-events-none" />
        <Input
          type="search"
          placeholder={t(($) => $.toolbar.search_placeholder)}
          value={state.searchQuery}
          onChange={(e) => onChange({ searchQuery: e.target.value })}
          className="h-7 pl-7 pr-6 text-xs"
          aria-label={t(($) => $.toolbar.search_label)}
        />
        {state.searchQuery && (
          <button
            type="button"
            onClick={onClearSearch}
            className="absolute right-1.5 top-1/2 -translate-y-1/2 rounded p-0.5 text-muted-foreground hover:text-foreground"
            aria-label={t(($) => $.toolbar.clear_search)}
          >
            <X className="h-3 w-3" />
          </button>
        )}
      </div>

      <div className="flex-1" />

      {/* Unread filter toggle */}
      <Tooltip>
        <TooltipTrigger
          render={
            <Toggle
              pressed={state.unreadOnly}
              onPressedChange={(v) => onChange({ unreadOnly: v })}
              variant="outline"
              size="sm"
              aria-label={t(($) => $.toolbar.unread_only)}
              className="h-7 px-2"
            >
              <Filter className="h-3.5 w-3.5" />
              {!state.unreadOnly && unreadCount > 0 && (
                <span className="ml-1 text-[10px] tabular-nums">{unreadCount}</span>
              )}
            </Toggle>
          }
        />
        <TooltipContent side="bottom">
          {t(($) => $.toolbar.unread_only)}
          {unreadCount > 0 ? ` (${unreadCount})` : ""}
        </TooltipContent>
      </Tooltip>

      {/* Group mode dropdown */}
      <DropdownMenu>
        <DropdownMenuTrigger
          render={
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2 gap-1 text-xs"
              aria-label={t(($) => $.toolbar.group_label)}
            >
              <LayoutGrid className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">{groupLabels[state.groupMode]}</span>
            </Button>
          }
        />
        <DropdownMenuContent align="end" className="w-44">
          <DropdownMenuLabel className="text-xs">
            {t(($) => $.toolbar.group_by)}
          </DropdownMenuLabel>
          <DropdownMenuSeparator />
          {(["none", "time", "severity", "project", "type"] as const).map(
            (mode) => (
              <DropdownMenuCheckboxItem
                key={mode}
                checked={state.groupMode === mode}
                onCheckedChange={() => onChange({ groupMode: mode })}
              >
                {groupLabels[mode]}
              </DropdownMenuCheckboxItem>
            ),
          )}
        </DropdownMenuContent>
      </DropdownMenu>

      {/* Density toggle */}
      <Tooltip>
        <TooltipTrigger
          render={
            <Button
              variant="outline"
              size="sm"
              className="h-7 px-2"
              aria-label={t(($) => $.toolbar.density_label)}
              onClick={() =>
                onChange({
                  density:
                    state.density === "comfortable" ? "compact" : "comfortable",
                })
              }
            >
              <ListFilter className="h-3.5 w-3.5" />
            </Button>
          }
        />
        <TooltipContent side="bottom">
          {state.density === "comfortable"
            ? t(($) => $.toolbar.compact_view)
            : t(($) => $.toolbar.comfortable_view)}
        </TooltipContent>
      </Tooltip>

      {/* Item count */}
      <span className="text-[10px] text-muted-foreground tabular-nums min-w-[2ch] text-right">
        {totalCount}
      </span>
    </div>
  );
}
