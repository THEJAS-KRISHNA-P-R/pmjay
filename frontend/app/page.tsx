import { Header } from "./components/Header";
import { IntakeForm } from "./components/IntakeForm";

export default function HomePage() {
  return (
    <div className="min-h-screen flex flex-col">
      <Header />
      <main className="mx-auto w-full max-w-2xl flex-1 px-4 py-8 sm:px-6 sm:py-12 animate-enter">
        <div className="space-y-3.5">
          <div className="inline-flex items-center rounded-full bg-teal-100/70 px-4 py-1.5 text-xs font-bold uppercase tracking-wider text-teal-800 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.7)] backdrop-blur-sm">
            Hospital Billing Assistance
          </div>
          
          <h1 className="font-display text-3xl font-semibold leading-tight tracking-tight-display text-teal-950 sm:text-4xl">
            Is the hospital right to deny your PMJAY claim?
          </h1>
          
          <p className="text-base sm:text-lg leading-relaxed tracking-body text-sand-900/80">
            Free, instant help figuring out whether a hospital&rsquo;s coverage denial is correct —
            right when you&rsquo;re standing at the billing desk. This is not a government service
            and it does not replace the official complaint process, but it can help you understand
            what&rsquo;s happening and what to say next.
          </p>
        </div>

        <div className="mt-8 glass-panel-elevated p-6 sm:p-8">
          <IntakeForm />
        </div>

        <div className="mt-8 rounded-2xl bg-white/70 p-5 text-center text-xs sm:text-sm text-sand-900/75 backdrop-blur-md shadow-xs">
          <p>
            In an emergency, get treatment first, always. Call the PMJAY helpline on{" "}
            <a 
              href="tel:14555" 
              className="font-bold text-teal-800 underline decoration-teal-500/50 hover:decoration-teal-700 transition touch-spring"
            >
              14555
            </a>{" "}
            any time.
          </p>
        </div>
      </main>
    </div>
  );
}
