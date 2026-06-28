"use client";

import { useT } from "../../i18n";

/** Hook returning a localized relative-time formatter. */
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
