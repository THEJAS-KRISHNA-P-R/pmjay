/**
 * localPrefs.ts — small local-only preferences, same rationale as
 * lib/caseHistory.ts (no accounts, no backend to store this on). Two
 * clearly different kinds of thing live here, both scoped honestly:
 *
 * - name/phone/email: nothing in this app currently *uses* these for
 *   anything (no complaint template inserts a name, no notification
 *   system exists to email or text). They're stored only so a family
 *   who wants to jot them down somewhere convenient can, without this
 *   pretending to be more than that. Settings says this plainly.
 * - preferredLanguage: honestly framed the same way — saved, and used
 *   to help prioritize which languages get full interface translation
 *   next, not something that changes today's interface language. See
 *   DESIGN.md and backend/internal/extract/prompt.go's language section
 *   for the actual, current state of language support.
 * - textScaleLarge: the one setting here that's fully real and
 *   functional today — see globals.css's html[data-text-scale="large"]
 *   rule and layout.tsx's pre-paint bootstrap script.
 */

const STORAGE_KEY = "pmjay-advocate:prefs:v1";
const TEXT_SCALE_KEY = "pmjay-advocate:text-scale";

export interface LocalPrefs {
  name: string;
  phone: string;
  email: string;
  preferredLanguage: string;
  textScaleLarge: boolean;
}

const DEFAULT_PREFS: LocalPrefs = {
  name: "",
  phone: "",
  email: "",
  preferredLanguage: "",
  textScaleLarge: false,
};

function isBrowser(): boolean {
  return typeof window !== "undefined" && typeof window.localStorage !== "undefined";
}

export function getLocalPrefs(): LocalPrefs {
  if (!isBrowser()) return { ...DEFAULT_PREFS };
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    const textScaleLarge = window.localStorage.getItem(TEXT_SCALE_KEY) === "large";
    if (!raw) return { ...DEFAULT_PREFS, textScaleLarge };
    const parsed = JSON.parse(raw);
    return { ...DEFAULT_PREFS, ...parsed, textScaleLarge };
  } catch {
    return { ...DEFAULT_PREFS };
  }
}

export function saveLocalPrefs(partial: Partial<Omit<LocalPrefs, "textScaleLarge">>): LocalPrefs {
  const current = getLocalPrefs();
  const next = { ...current, ...partial };
  if (isBrowser()) {
    try {
      const toStore = { name: next.name, phone: next.phone, email: next.email, preferredLanguage: next.preferredLanguage };
      window.localStorage.setItem(STORAGE_KEY, JSON.stringify(toStore));
    } catch {
      // Best-effort, same posture as caseHistory.ts.
    }
  }
  return next;
}

export function setTextScaleLarge(large: boolean): void {
  if (!isBrowser()) return;
  try {
    if (large) {
      window.localStorage.setItem(TEXT_SCALE_KEY, "large");
      document.documentElement.setAttribute("data-text-scale", "large");
    } else {
      window.localStorage.removeItem(TEXT_SCALE_KEY);
      document.documentElement.removeAttribute("data-text-scale");
    }
  } catch {
    // Best-effort.
  }
}

/** Clears saved name/phone/email/language — never touches case history
 * (lib/caseHistory.ts's clearAllHistory is the separate, explicit
 * action for that) or the text-scale accessibility setting, since
 * "forget my contact details" and "turn off large text" are different
 * intents that shouldn't be bundled into one button by surprise. */
export function clearLocalPrefs(): void {
  if (!isBrowser()) return;
  try {
    window.localStorage.removeItem(STORAGE_KEY);
  } catch {
    // Best-effort.
  }
}
