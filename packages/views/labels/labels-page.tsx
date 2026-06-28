"use client";

import { PageHeader } from "../layout/page-header";
import { LabelsPanel } from "../issues/components/labels-panel";
import { useT } from "../i18n";

/**
 * Standalone labels management page. Reuses the existing LabelsPanel
 * (previously only accessible through the LabelPicker dialog on the
 * issue detail page) inside a full-page layout with a header bar.
 */
export function LabelsPage() {
  const { t } = useT("labels");
  return (
    <div className="flex flex-col h-full">
      <PageHeader>
        <h1 className="text-sm font-medium">{t(($) => $.page.title)}</h1>
      </PageHeader>
      <div className="flex-1 overflow-y-auto p-6">
        <div className="max-w-2xl">
          <LabelsPanel />
        </div>
      </div>
    </div>
  );
}
