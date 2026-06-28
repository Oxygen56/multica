import { useEffect, useCallback } from "react";
import type { InboxItem } from "@multica/core/types";

export interface KeyboardNavState {
  /** Index of the currently focused item in the flat list, or -1 if none */
  focusedIndex: number;
  /** Whether multi-select mode is active */
  multiSelectMode: boolean;
  /** Set of selected item IDs */
  selectedIds: Set<string>;
}

export interface KeyboardNavActions {
  setFocusedIndex: (i: number) => void;
  moveFocusUp: () => void;
  moveFocusDown: () => void;
  archiveFocused: () => void;
  focusSearch: () => void;
  toggleMultiSelect: () => void;
  toggleSelectFocused: () => void;
}

/**
 * Hook that binds keyboard shortcuts for inbox navigation.
 *
 * Shortcuts:
 *   j / ArrowDown  — move focus down
 *   k / ArrowUp    — move focus up
 *   e              — archive focused item
 *   /              — focus the search input
 *   Esc            — clear search / exit multi-select
 *   x              — toggle selection of focused item (in multi-select mode)
 *   Ctrl+A         — select all (in multi-select mode)
 */
export function useInboxKeyboardNav(
  items: InboxItem[],
  navState: KeyboardNavState,
  actions: KeyboardNavActions,
  enabled: boolean,
) {
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      // Don't capture when focus is in an input, textarea, or contenteditable
      const target = e.target as HTMLElement;
      const tag = target.tagName.toLowerCase();
      const isEditable =
        tag === "input" || tag === "textarea" || target.isContentEditable;
      if (isEditable) return;

      // Allow escape to always work from search input
      const isSearchInput =
        tag === "input" && (target as HTMLInputElement).type === "search";

      switch (e.key) {
        case "j":
        case "ArrowDown": {
          e.preventDefault();
          if (navState.focusedIndex < items.length - 1) {
            actions.moveFocusDown();
          }
          break;
        }
        case "k":
        case "ArrowUp": {
          e.preventDefault();
          if (navState.focusedIndex > 0) {
            actions.moveFocusUp();
          }
          break;
        }
        case "e": {
          if (isEditable) break;
          if (navState.multiSelectMode) {
            e.preventDefault();
            actions.toggleSelectFocused();
          } else if (navState.focusedIndex >= 0) {
            e.preventDefault();
            actions.archiveFocused();
          }
          break;
        }
        case "/": {
          if (isEditable) break;
          e.preventDefault();
          actions.focusSearch();
          break;
        }
        case "Escape": {
          if (isSearchInput) {
            // Blur the search input so a second Escape triggers clear/exit
            (target as HTMLInputElement).blur();
            break;
          }
          e.preventDefault();
          if (navState.multiSelectMode) {
            actions.toggleMultiSelect();
          }
          break;
        }
        case "x": {
          if (isEditable) break;
          if (navState.multiSelectMode && navState.focusedIndex >= 0) {
            e.preventDefault();
            actions.toggleSelectFocused();
          }
          break;
        }
        case "a": {
          if (e.ctrlKey || e.metaKey) {
            if (navState.multiSelectMode && !isEditable) {
              e.preventDefault();
              // Select all — handled by the page component
            }
          }
          break;
        }
      }
    },
    [items.length, navState, actions],
  );

  useEffect(() => {
    if (!enabled) return;
    document.addEventListener("keydown", handleKeyDown);
    return () => document.removeEventListener("keydown", handleKeyDown);
  }, [enabled, handleKeyDown]);
}
