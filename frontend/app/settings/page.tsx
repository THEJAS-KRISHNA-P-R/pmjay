"use client";

import { useEffect, useState } from "react";
import { AppShell } from "@/app/components/AppShell";
import { getLocalPrefs, saveLocalPrefs, setTextScaleLarge, clearLocalPrefs } from "@/lib/localPrefs";
import { listCaseHistory, clearAllHistory } from "@/lib/caseHistory";
import type { LocalPrefs } from "@/lib/localPrefs";
import { IconUser, IconGlobe, IconBell, IconLock, IconTrash, IconCheck, IconEye } from "@/app/components/icons";

const LANGUAGE_OPTIONS = [
  "English",
  "Hindi / हिन्दी",
  "Malayalam / മലയാളം",
  "Tamil / தமிழ்",
  "Telugu / తెలుగు",
  "Kannada / ಕನ್ನಡ",
  "Bengali / বাংলা",
  "Marathi / मराठी",
  "Gujarati / ગુજરાતી",
  "Punjabi / ਪੰਜਾਬੀ",
];

export default function SettingsPage() {
  const [prefs, setPrefs] = useState<LocalPrefs | null>(null);
  const [errors, setErrors] = useState<{ phone?: string; email?: string }>({});
  const [caseCount, setCaseCount] = useState(0);
  const [savedFlash, setSavedFlash] = useState(false);
  const [clearedCases, setClearedCases] = useState(false);
  const [clearedPrefs, setClearedPrefs] = useState(false);

  useEffect(() => {
    setPrefs(getLocalPrefs());
    setCaseCount(listCaseHistory().length);
  }, []);

  function validatePhone(raw: string): string | undefined {
    const trimmed = raw.trim();
    if (!trimmed) return undefined; // Optional field
    const digits = trimmed.replace(/\D/g, "");
    if (digits.length === 10 && /^[6-9]\d{9}$/.test(digits)) {
      return undefined;
    }
    if (digits.length === 11 && /^0[6-9]\d{9}$/.test(digits)) {
      return undefined;
    }
    if (digits.length === 12 && /^91[6-9]\d{9}$/.test(digits)) {
      return undefined;
    }
    return "Please enter a valid 10-digit Indian mobile number (e.g. 9876543210 or +91 98765 43210)";
  }

  function validateEmail(raw: string): string | undefined {
    const trimmed = raw.trim();
    if (!trimmed) return undefined; // Optional field
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(trimmed)) {
      return "Please enter a valid email address (e.g. name@example.com)";
    }
    return undefined;
  }

  function updateField<K extends keyof Omit<LocalPrefs, "textScaleLarge">>(key: K, value: LocalPrefs[K]) {
    if (!prefs) return;
    setPrefs({ ...prefs, [key]: value });
    if (errors[key as keyof typeof errors]) {
      setErrors((prev) => ({ ...prev, [key]: undefined }));
    }
  }

  function handlePhoneChange(val: string) {
    // Only allow digits, plus, hyphens, and spaces while typing
    const sanitized = val.replace(/[^0-9+\s-]/g, "");
    updateField("phone", sanitized);
  }

  function handleSave(e: React.FormEvent) {
    e.preventDefault();
    if (!prefs) return;

    const phoneErr = validatePhone(prefs.phone);
    const emailErr = validateEmail(prefs.email);

    if (phoneErr || emailErr) {
      setErrors({
        phone: phoneErr,
        email: emailErr,
      });
      return;
    }

    setErrors({});
    saveLocalPrefs({
      name: prefs.name.trim(),
      phone: prefs.phone.trim(),
      email: prefs.email.trim(),
      preferredLanguage: prefs.preferredLanguage,
    });
    setSavedFlash(true);
    setTimeout(() => setSavedFlash(false), 2000);
  }

  function handleToggleTextScale() {
    if (!prefs) return;
    const next = !prefs.textScaleLarge;
    setTextScaleLarge(next);
    setPrefs({ ...prefs, textScaleLarge: next });
  }

  function handleClearCases() {
    clearAllHistory();
    setCaseCount(0);
    setClearedCases(true);
  }

  function handleClearPrefs() {
    clearLocalPrefs();
    setPrefs(getLocalPrefs());
    setClearedPrefs(true);
  }

  if (!prefs) {
    return (
      <AppShell>
        <div className="space-y-4" role="status" aria-label="Loading settings">
          <div className="h-8 w-32 rounded-lg skeleton-shimmer" />
          <div className="h-64 rounded-2xl skeleton-shimmer" />
        </div>
      </AppShell>
    );
  }

  return (
    <AppShell>
      <div className="w-full max-w-6xl space-y-6">
        <div className="space-y-1.5">
          <h1 className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-ink-950">
            Settings
          </h1>
          <p className="text-sm sm:text-base text-sand-600 font-medium">
            There&rsquo;s no account here: everything below is saved only in this browser.
          </p>
        </div>

        <div className="grid lg:grid-cols-12 gap-6 lg:gap-8 items-start">
          {/* Left Column: Form Details & Language */}
          <div className="lg:col-span-7 space-y-6">
            <form onSubmit={handleSave} className="card p-6 sm:p-8 space-y-5 shadow-sm">
              <SectionHeading Icon={IconUser} title="Your details" />
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed -mt-2">
                Nothing in this tool uses these yet (no complaint template fills them in automatically, and
                nothing sends you a notification). This is just a convenient place to keep them for yourself.
              </p>
              <div className="grid sm:grid-cols-2 gap-4">
                <label className="block text-sm font-bold text-sand-800">
                  Name
                  <input
                    type="text"
                    value={prefs.name}
                    onChange={(e) => updateField("name", e.target.value)}
                    maxLength={60}
                    className="field mt-1.5 font-normal"
                    placeholder="Your name"
                  />
                </label>
                <label className="block text-sm font-bold text-sand-800">
                  Phone
                  <input
                    type="tel"
                    value={prefs.phone}
                    onChange={(e) => handlePhoneChange(e.target.value)}
                    aria-invalid={Boolean(errors.phone)}
                    aria-describedby={errors.phone ? "phone-error" : undefined}
                    className={`field mt-1.5 font-normal ${
                      errors.phone ? "border-tier-red-border focus:border-tier-red-text focus:ring-2 focus:ring-tier-red-bg" : ""
                    }`}
                    placeholder="e.g. 9876543210 or +91 98765 43210"
                  />
                  {errors.phone && (
                    <p id="phone-error" role="alert" className="mt-1.5 text-xs font-bold text-tier-red-text animate-fade-in">
                      {errors.phone}
                    </p>
                  )}
                </label>
                <label className="block text-sm font-bold text-sand-800 sm:col-span-2">
                  Email
                  <input
                    type="email"
                    value={prefs.email}
                    onChange={(e) => updateField("email", e.target.value)}
                    aria-invalid={Boolean(errors.email)}
                    aria-describedby={errors.email ? "email-error" : undefined}
                    className={`field mt-1.5 font-normal ${
                      errors.email ? "border-tier-red-border focus:border-tier-red-text focus:ring-2 focus:ring-tier-red-bg" : ""
                    }`}
                    placeholder="e.g. yourname@example.com"
                  />
                  {errors.email && (
                    <p id="email-error" role="alert" className="mt-1.5 text-xs font-bold text-tier-red-text animate-fade-in">
                      {errors.email}
                    </p>
                  )}
                </label>
              </div>

              <div className="pt-3 border-t border-sand-100 space-y-3">
                <SectionHeading Icon={IconGlobe} title="Preferred language" />
                <p className="text-xs sm:text-sm text-sand-600 leading-relaxed -mt-2">
                  You can describe your situation in any language regardless of what you pick here. This
                  setting does not restrict input, but helps us prioritize full interface translations next.
                </p>
                <select
                  value={prefs.preferredLanguage}
                  onChange={(e) => updateField("preferredLanguage", e.target.value)}
                  className="w-full rounded-xl border border-sand-200 bg-white px-3 py-2.5 text-sm text-sand-900 placeholder:text-sand-400 focus:outline-none focus:ring-2 focus:ring-emerald-500/30 focus:border-emerald-500 transition-colors"
                >
                  <option value="">No preference</option>
                  {LANGUAGE_OPTIONS.map((lang) => (
                    <option key={lang} value={lang}>
                      {lang}
                    </option>
                  ))}
                </select>
              </div>

              <div className="pt-2 flex items-center gap-3">
                <button type="submit" className="btn-primary tap-target px-6 py-3 text-sm">
                  Save
                </button>
                {savedFlash && (
                  <span className="inline-flex items-center gap-1.5 text-sm font-bold text-tier-green-text animate-fade-in">
                    <IconCheck className="h-4 w-4" /> Saved
                  </span>
                )}
              </div>
            </form>
          </div>

          {/* Right Column: Accessibility, Notifications & Data Management */}
          <div className="lg:col-span-5 space-y-6">
            <div className="card p-6 sm:p-7 space-y-4 shadow-sm">
              <SectionHeading Icon={IconEye} title="Accessibility" />
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="text-sm font-bold text-sand-900">Larger text</p>
                  <p className="text-xs sm:text-sm text-sand-600 mt-0.5">
                    Scales all text in the app up, right away, on this device.
                  </p>
                </div>
                <button
                  type="button"
                  role="switch"
                  aria-checked={prefs.textScaleLarge}
                  onClick={handleToggleTextScale}
                  aria-label="Toggle larger text mode"
                  className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer items-center rounded-full transition-colors duration-200 ease-in-out focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ink-600 focus-visible:ring-offset-2 ${
                    prefs.textScaleLarge ? "bg-teal-600" : "bg-sand-300"
                  }`}
                >
                  <span
                    className={`pointer-events-none inline-block h-5 w-5 transform rounded-full bg-white shadow-sm ring-0 transition duration-200 ease-in-out ${
                      prefs.textScaleLarge ? "translate-x-5.5" : "translate-x-0.5"
                    }`}
                  />
                </button>
              </div>
            </div>

            <div className="card p-6 sm:p-7 space-y-3 shadow-sm">
              <SectionHeading Icon={IconBell} title="Notifications" />
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed">
                There&rsquo;s nothing to configure yet. This tool doesn&rsquo;t send emails, texts, or push
                notifications of any kind. If that changes, this is where you&rsquo;d control it.
              </p>
            </div>

            <div className="card p-6 sm:p-7 space-y-4 shadow-sm">
              <SectionHeading Icon={IconLock} title="Privacy &amp; data" />
              <p className="text-xs sm:text-sm text-sand-600 leading-relaxed">
                Your case{caseCount === 1 ? "" : "s"} ({caseCount} tracked on this device) and saved details are
                stored only in this browser&rsquo;s local storage.
              </p>

              <div className="flex flex-col gap-2.5 pt-1">
                <button type="button" onClick={handleClearCases} className="btn-secondary tap-target px-4 py-2.5 text-xs font-bold justify-start">
                  <IconTrash className="h-4 w-4" />
                  <span>Clear this browser&rsquo;s case list</span>
                </button>
                <button type="button" onClick={handleClearPrefs} className="btn-secondary tap-target px-4 py-2.5 text-xs font-bold justify-start">
                  <IconTrash className="h-4 w-4" />
                  <span>Clear saved details</span>
                </button>
              </div>
              <div aria-live="polite" className="space-y-1">
                {clearedCases && <p className="text-xs font-bold text-tier-green-text">Case list cleared on this device.</p>}
                {clearedPrefs && <p className="text-xs font-bold text-tier-green-text">Saved details cleared.</p>}
              </div>
            </div>
          </div>
        </div>
      </div>
    </AppShell>
  );
}

function SectionHeading({ Icon, title }: { Icon: typeof IconUser; title: string }) {
  return (
    <div className="flex items-center gap-2.5">
      <div className="flex h-7 w-7 items-center justify-center rounded-lg bg-sand-100 border border-sand-200/60 text-sand-700 shadow-[inset_0_1px_1px_rgba(255,255,255,0.9),0_1px_2px_rgba(42,38,33,0.04)]">
        <Icon className="h-4 w-4" />
      </div>
      <h2 className="text-base sm:text-lg font-bold tracking-tight text-sand-900">{title}</h2>
    </div>
  );
}
