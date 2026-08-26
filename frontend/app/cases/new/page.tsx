"use client";

import { AppShell } from "@/app/components/AppShell";
import { IntakeForm } from "@/app/components/IntakeForm";
import { IconLock, IconClock, IconShieldCheck } from "@/app/components/icons";

export default function NewCasePage() {
  return (
    <AppShell>
      <div className="w-full max-w-6xl space-y-6">
        <div className="space-y-1.5">
          <p className="text-xs font-bold uppercase tracking-wider text-emerald-700">New case verification</p>
          <h1 className="font-display text-2xl sm:text-3xl font-semibold tracking-tight-display text-ink-950">
            Tell us what happened
          </h1>
          <p className="text-sm sm:text-base text-sand-600 font-medium leading-relaxed">
            Take your time, and use your own words (this becomes a case you can come back to).
          </p>
        </div>

        <div className="grid lg:grid-cols-12 gap-6 lg:gap-8 items-start">
          {/* Main Form Column */}
          <div className="lg:col-span-7 space-y-6">
            <div className="card p-6 sm:p-8 shadow-sm">
              <IntakeForm />
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <TrustPoint Icon={IconClock} text="Usually answers in under a minute" />
              <TrustPoint Icon={IconLock} text="No login required: stays on your device" />
              <TrustPoint Icon={IconShieldCheck} text="Free to use, always" />
            </div>
          </div>

          {/* Right Guidance & Information Column */}
          <div className="lg:col-span-5 space-y-5">
            <div className="card p-6 border border-sand-200/80 bg-white space-y-3.5 shadow-xs">
              <h3 className="text-sm font-bold text-sand-900 flex items-center gap-2">
                <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-emerald-50 text-emerald-700">
                  <IconShieldCheck className="h-4 w-4" />
                </span>
                What helps us check your bill
              </h3>
              <ul className="space-y-3 text-xs text-sand-600 font-medium leading-relaxed">
                <li className="flex items-start gap-2">
                  <span className="text-emerald-700 font-bold">•</span>
                  <span><strong>The procedure or symptom:</strong> Name of surgery, test, or department (e.g. angioplasty, fracture, gallstone).</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-emerald-700 font-bold">•</span>
                  <span><strong>What the hospital asked for:</strong> Cash deposit, medicine payment outside, or verbal refusal.</span>
                </li>
                <li className="flex items-start gap-2">
                  <span className="text-emerald-700 font-bold">•</span>
                  <span><strong>Any language:</strong> English, Hindi, Malayalam, or mixed sentences are all fully supported.</span>
                </li>
              </ul>
            </div>

            <div className="card p-6 border border-emerald-200/60 bg-emerald-50/50 space-y-3 shadow-xs">
              <h3 className="text-sm font-bold text-emerald-950 flex items-center gap-2">
                <span className="flex h-6 w-6 items-center justify-center rounded-lg bg-emerald-100 text-emerald-800">
                  <IconLock className="h-3.5 w-3.5" />
                </span>
                How this tool protects you
              </h3>
              <p className="text-xs text-emerald-900 leading-relaxed font-medium">
                We compare the hospital&rsquo;s request directly against the 315+ official packages in the National Health Authority (NHA) master schedule. If the hospital is empanelled, you are guaranteed 100% cashless treatment with ₹0 out-of-pocket charges.
              </p>
            </div>
          </div>
        </div>
      </div>
    </AppShell>
  );
}

function TrustPoint({ Icon, text }: { Icon: typeof IconLock; text: string }) {
  return (
    <div className="card flex items-center gap-2.5 px-3.5 py-3">
      <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-lg bg-ink-50 text-ink-700">
        <Icon className="h-3.5 w-3.5" />
      </span>
      <p className="text-xs font-bold text-sand-700 leading-snug">{text}</p>
    </div>
  );
}
