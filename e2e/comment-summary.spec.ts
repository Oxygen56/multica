import { test, expect } from "@playwright/test";
import { createTestApi, loginAsDefault, waitForPageText } from "./helpers";
import type { TestApiClient } from "./fixtures";

for (const isReply of [false, true]) {
  for (const expandFirst of [false, true]) {
    test(`summary ${isReply ? "reply" : "root"} preserves full text when editing ${expandFirst ? "after expansion" : "directly"}`, async ({ page }) => {
      let api: TestApiClient | undefined;
      try {
        api = await createTestApi();
        const title = "E2E long comment " + Date.now();
        const issue = await api.createIssue(title);
        const slug = await loginAsDefault(page);
        const content = "Comment prefix " + "long body ".repeat(35) + "TAIL-MUST-SURVIVE";
        const base = process.env.NEXT_PUBLIC_API_URL || `http://localhost:${process.env.PORT}`;
        const headers = { "Content-Type": "application/json", Authorization: `Bearer ${api.getToken()}`, "X-Workspace-Slug": slug };
        let parentId: string | undefined;
        if (isReply) {
          const parent = await fetch(`${base}/api/issues/${issue.id}/comments`, { method: "POST", headers, body: JSON.stringify({ content: "Parent comment" }) });
          expect(parent.ok).toBe(true);
          parentId = (await parent.json()).id;
        }
        const response = await fetch(`${base}/api/issues/${issue.id}/comments`, { method: "POST", headers, body: JSON.stringify({ content, parent_id: parentId }) });
        expect(response.ok).toBe(true);
        const comment = await response.json();
        const timelineResponse = page.waitForResponse(r => r.url().includes(`/api/issues/${issue.id}/timeline?summary=true`));
        await page.goto(`/${slug}/issues/${issue.id}`, { waitUntil: "domcontentloaded" });
        // The issue route canonicalizes UUIDs to workspace issue keys. Wait
        // for that navigation before opening menus that the remount replaces.
        await page.waitForURL(url => url.pathname.includes("/issues/") && !url.pathname.endsWith(issue.id));
        await waitForPageText(page, title);
        const entries = await (await timelineResponse).json();
        const summary = entries.find((e: { id: string }) => e.id === comment.id);
        expect(summary.content_truncated).toBe(true);
        expect(summary.content).not.toContain("TAIL-MUST-SURVIVE");
        if (expandFirst) {
          if (!isReply) await page.screenshot({ path: test.info().outputPath("summary.png") });
          await page.getByRole("button", { name: "Expand full text", exact: true }).click();
          await expect(page.getByText("TAIL-MUST-SURVIVE", { exact: false })).toBeVisible();
          if (!isReply) await page.screenshot({ path: test.info().outputPath("expanded.png") });
        }
        const row = page.locator(`[data-comment-content="${comment.id}"]`).locator("xpath=ancestor::*[@data-comment-block][1]");
        // Comment actions appear after hovering their owning block.
        await row.hover();
        const actions = row.getByRole("button", { name: "Comment actions", exact: true });
        await actions.click();
        const editAction = page.getByRole("menuitem", { name: "Edit", exact: true });
        await editAction.click();
        const editor = page.locator(".ProseMirror").filter({ hasText: "Comment prefix" });
        await expect(editor).toContainText("TAIL-MUST-SURVIVE");
        await editor.fill(content + " UPDATED");
        const saved = page.waitForResponse(r => r.url().endsWith(`/api/comments/${comment.id}`) && r.request().method() === "PUT");
        await page.getByRole("button", { name: "Save", exact: true }).click();
        expect((await saved).ok()).toBe(true);
        const detail = await fetch(`${base}/api/comments/${comment.id}`, { headers });
        expect(detail.ok).toBe(true);
        expect((await detail.json()).content).toBe(content + " UPDATED");
      } finally {
        await api?.cleanup();
      }
    });
  }
}
