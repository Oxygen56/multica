package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestListTimeline_SummaryAndFullComment(t *testing.T) {
	issueID := createIssueForTimeline(t, "Timeline summary")
	comments, _ := seedTimelineEntries(t, issueID, 1, 1)
	content := strings.Repeat("正文", 120) + "END"
	if _, err := testPool.Exec(context.Background(), `UPDATE comment SET content = $1 WHERE id = $2`, content, comments[0]); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"?summary=true", "?summary=true&limit=10"} {
		w := httptest.NewRecorder()
		req := withURLParam(newRequest("GET", "/api/issues/"+issueID+"/timeline"+suffix, nil), "id", issueID)
		testHandler.ListTimeline(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("timeline status %d: %s", w.Code, w.Body)
		}
		var entries []TimelineEntry
		if strings.Contains(suffix, "limit") {
			var page timelinePaginatedResponse
			if err := json.Unmarshal(w.Body.Bytes(), &page); err != nil {
				t.Fatal(err)
			}
			entries = page.Entries
		} else if err := json.Unmarshal(w.Body.Bytes(), &entries); err != nil {
			t.Fatal(err)
		}
		found := false
		for _, entry := range entries {
			if entry.ID == comments[0] {
				found = true
				if entry.ContentTruncated == nil || !*entry.ContentTruncated || entry.Content == nil || utf8.RuneCountInString(*entry.Content) != summaryContentRunes+1 {
					t.Fatalf("bad summary: %+v", entry)
				}
			}
			if entry.Type == "activity" && entry.ContentTruncated != nil {
				t.Fatal("activity must not be summarized")
			}
		}
		if !found {
			t.Fatal("comment missing")
		}
	}
	entries, status := fetchTimeline(t, issueID)
	if status != http.StatusOK {
		t.Fatalf("full timeline status %d", status)
	}
	for _, entry := range entries {
		if entry.ID == comments[0] && (entry.Content == nil || *entry.Content != content || entry.ContentTruncated != nil) {
			t.Fatalf("default lost content: %+v", entry)
		}
	}
	w := httptest.NewRecorder()
	req := withURLParam(newRequest("GET", "/api/comments/"+comments[0], nil), "commentId", comments[0])
	testHandler.GetComment(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("detail status %d: %s", w.Code, w.Body)
	}
	var full CommentResponse
	if err := json.Unmarshal(w.Body.Bytes(), &full); err != nil {
		t.Fatal(err)
	}
	if full.Content != content || full.ID != comments[0] {
		t.Fatalf("detail lost content: %+v", full)
	}
}
