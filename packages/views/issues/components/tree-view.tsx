"use client";

import { useState, useCallback, useMemo } from "react";
import { ChevronRight } from "lucide-react";
import type { Issue } from "@multica/core/types";
import { useT } from "../../i18n";
import type { ChildProgress } from "./list-row";
import { TreeRow } from "./tree-row";

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

interface TreeNode {
  issue: Issue;
  children: TreeNode[];
}

// ---------------------------------------------------------------------------
// Tree building
// ---------------------------------------------------------------------------

function buildTree(issues: Issue[]): { roots: TreeNode[]; nodeMap: Map<string, TreeNode> } {
  const nodeMap = new Map<string, TreeNode>();
  const roots: TreeNode[] = [];

  // First pass: create a node for every issue
  for (const issue of issues) {
    nodeMap.set(issue.id, { issue, children: [] });
  }

  // Second pass: attach children to their parents
  for (const issue of issues) {
    const node = nodeMap.get(issue.id)!;
    if (issue.parent_issue_id) {
      const parent = nodeMap.get(issue.parent_issue_id);
      if (parent) {
        parent.children.push(node);
      } else {
        // Parent not in the current list — treat as root
        roots.push(node);
      }
    } else {
      roots.push(node);
    }
  }

  return { roots, nodeMap };
}

// ---------------------------------------------------------------------------
// TreeView
// ---------------------------------------------------------------------------

export function TreeView({
  issues,
  childProgressMap,
}: {
  issues: Issue[];
  childProgressMap: Map<string, ChildProgress>;
}) {
  const { t } = useT("issues");

  const { roots } = useMemo(() => buildTree(issues), [issues]);

  const [collapsedIds, setCollapsedIds] = useState<Set<string>>(new Set());

  const toggleCollapsed = useCallback((id: string) => {
    setCollapsedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  }, []);

  if (roots.length === 0) {
    return (
      <div className="flex-1 min-h-0 flex items-center justify-center p-4 text-sm text-muted-foreground">
        {t(($) => $.list.empty_status)}
      </div>
    );
  }

  return (
    <div className="flex-1 min-h-0 overflow-y-auto p-2 pt-0">
      <div className="space-y-0">
        {roots.map((node) => (
          <TreeNodeRow
            key={node.issue.id}
            node={node}
            depth={0}
            collapsedIds={collapsedIds}
            onToggleCollapse={toggleCollapsed}
            childProgressMap={childProgressMap}
          />
        ))}
      </div>
    </div>
  );
}

// ---------------------------------------------------------------------------
// TreeNodeRow (recursive)
// ---------------------------------------------------------------------------

function TreeNodeRow({
  node,
  depth,
  collapsedIds,
  onToggleCollapse,
  childProgressMap,
}: {
  node: TreeNode;
  depth: number;
  collapsedIds: Set<string>;
  onToggleCollapse: (id: string) => void;
  childProgressMap: Map<string, ChildProgress>;
}) {
  const hasChildren = node.children.length > 0;
  const isCollapsed = collapsedIds.has(node.issue.id);
  const childProgress = childProgressMap.get(node.issue.id);
  // Indent by 20px per depth level, with a minimum of 0
  const indentPadding = depth * 20;

  return (
    <>
      <div className="flex items-center group/tree-row">
        {/* Indentation spacer */}
        {depth > 0 && (
          <div
            className="shrink-0"
            style={{ width: indentPadding }}
            aria-hidden="true"
          />
        )}

        {/* Expand/collapse toggle */}
        <div className="shrink-0 w-5 flex items-center justify-center">
          {hasChildren ? (
            <button
              type="button"
              className="flex items-center justify-center size-5 rounded-sm text-muted-foreground hover:text-foreground hover:bg-accent transition-colors cursor-pointer"
              onClick={() => onToggleCollapse(node.issue.id)}
              aria-label={isCollapsed ? "Expand sub-issues" : "Collapse sub-issues"}
            >
              <ChevronRight
                className={`size-3.5 transition-transform ${
                  isCollapsed ? "" : "rotate-90"
                }`}
              />
            </button>
          ) : (
            /* Empty spacer to keep alignment */
            <span className="size-5" />
          )}
        </div>

        {/* The actual row content */}
        <div className="flex-1 min-w-0">
          <TreeRow
            issue={node.issue}
            childProgress={childProgress}
          />
        </div>
      </div>

      {/* Children (recursive) */}
      {hasChildren &&
        !isCollapsed &&
        node.children.map((child) => (
          <TreeNodeRow
            key={child.issue.id}
            node={child}
            depth={depth + 1}
            collapsedIds={collapsedIds}
            onToggleCollapse={onToggleCollapse}
            childProgressMap={childProgressMap}
          />
        ))}
    </>
  );
}
