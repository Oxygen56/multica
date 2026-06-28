import { Extension } from "@tiptap/core";

/**
 * `onSubmit` must return true when it actually handled the event and false
 * when there's no submit handler wired up. That lets us fall through to the
 * default Enter behaviour — inserting a newline — when appropriate.
 *
 * `submitOnEnter` — when true, bare Enter submits (chat-style) and all
 * modifier combos (Shift-Enter, Mod-Enter) insert a newline. When false,
 * only Mod-Enter submits and bare Enter keeps its default (newline).
 */
export function createSubmitExtension(
  onSubmit: () => boolean,
  { submitOnEnter }: { submitOnEnter: boolean },
) {
  return Extension.create({
    name: "submitShortcut",
    addKeyboardShortcuts() {
      const shortcuts: Record<string, () => boolean> = {};
      if (submitOnEnter) {
        // Chat-style: bare Enter submits. All modifier + Enter combos
        // (Shift-Enter, Mod-Enter) are explicitly short-circuited to
        // return false so ProseMirror inserts a newline instead.
        shortcuts.Enter = () => {
          const editor = this.editor;
          // IME guard — never submit while composing a multi-key input
          // (Chinese pinyin, Japanese kana, etc). `view.composing` is set
          // by ProseMirror between compositionstart and compositionend.
          if (editor.view.composing) return false;
          // Let Enter insert a newline inside a code block.
          if (editor.isActive("codeBlock")) return false;
          return onSubmit();
        };
        shortcuts["Shift-Enter"] = () => false;
        shortcuts["Mod-Enter"] = () => false;
      } else {
        // Default: only Mod-Enter submits, bare Enter is a normal newline.
        shortcuts["Mod-Enter"] = () => onSubmit();
      }
      return shortcuts;
    },
  });
}
