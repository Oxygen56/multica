"use client";

import { CheckCheck, Archive, X, CheckSquare, Square } from "lucide-react";
import { Button } from "@multica/ui/components/ui/button";
import { Checkbox } from "@multica/ui/components/ui/checkbox";
import { useT } from "../../i18n";

export function InboxBatchBar({
  selectedIds,
  totalCount,
  onToggleSelectAll,
  onMarkReadSelected,
  onArchiveSelected,
  onClearSelection,
}: {
  selectedIds: Set<string>;
  totalCount: number;
  onToggleSelectAll: () => void;
  onMarkReadSelected: () => void;
  onArchiveSelected: () => void;
  onClearSelection: () => void;
}) {
  const { t } = useT("inbox");
  const count = selectedIds.size;
  const allSelected = count === totalCount && totalCount > 0;

  if (count === 0) return null;

  return (
    <div
      className="flex items-center gap-2 px-3 py-2 border-t shrink-0 bg-accent/50"
      role="toolbar"
      aria-label={t(($) => $.batch.aria_label)}
    >
      <Button
        variant="ghost"
        size="sm"
        className="h-7 gap-1.5 text-xs"
        onClick={onToggleSelectAll}
      >
        {allSelected ? (
          <Square className="h-3.5 w-3.5" />
        ) : (
          <CheckSquare className="h-3.5 w-3.5" />
        )}
        {allSelected
          ? t(($) => $.batch.deselect_all)
          : t(($) => $.batch.select_all)}
      </Button>

      <span className="text-xs text-muted-foreground mx-1">
        {t(($) => $.batch.selected_count, { count })}
      </span>

      <div className="flex-1" />

      <Button
        variant="outline"
        size="sm"
        className="h-7 gap-1.5 text-xs"
        onClick={onMarkReadSelected}
        disabled={count === 0}
      >
        <CheckCheck className="h-3.5 w-3.5" />
        {t(($) => $.batch.mark_read)}
      </Button>

      <Button
        variant="outline"
        size="sm"
        className="h-7 gap-1.5 text-xs"
        onClick={onArchiveSelected}
        disabled={count === 0}
      >
        <Archive className="h-3.5 w-3.5" />
        {t(($) => $.batch.archive)}
      </Button>

      <Button
        variant="ghost"
        size="icon-sm"
        className="h-7 w-7"
        onClick={onClearSelection}
        aria-label={t(($) => $.batch.clear_selection)}
      >
        <X className="h-3.5 w-3.5" />
      </Button>
    </div>
  );
}

/**
 * Checkbox cell rendered on each inbox row in batch-select mode.
 */
export function InboxSelectCheckbox({
  id,
  checked,
  onChange,
}: {
  id: string;
  checked: boolean;
  onChange: (id: string, checked: boolean) => void;
}) {
  return (
    <div className="flex items-center shrink-0 pl-3">
      <Checkbox
        checked={checked}
        onCheckedChange={(v) => onChange(id, v === true)}
        aria-label={`Select notification ${id}`}
        className="h-4 w-4"
      />
    </div>
  );
}
