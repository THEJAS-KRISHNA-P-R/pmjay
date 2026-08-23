import { Header } from "./components/Header";
import { IntakeForm } from "./components/IntakeForm";

export default function HomePage() {
  return (
    <div className="min-h-screen">
      <Header />
      <main className="mx-auto max-w-2xl px-4 py-8 sm:px-6 sm:py-12">
        <h1 className="font-display text-3xl font-semibold leading-tight text-teal-900 sm:text-4xl">
          Is the hospital right to deny your PMJAY claim?
        </h1>
        <p className="mt-3 text-lg leading-relaxed text-sand-900/80">
          Free, instant help figuring out whether a hospital&rsquo;s coverage denial is correct —
          right when you&rsquo;re standing at the billing desk. This is not a government service
          and it does not replace the official complaint process, but it can help you understand
          what&rsquo;s happening and what to say next.
        </p>

        <div className="mt-8 rounded-xl border border-sand-200 bg-white p-5 sm:p-6">
          <IntakeForm />
        </div>

        <p className="mt-6 text-center text-sm text-sand-900/60">
          In an emergency, get treatment first, always. Call the PMJAY helpline on{" "}
          <a href="tel:14555" className="font-bold text-teal-700 underline">
            14555
          </a>{" "}
          any time.
        </p>
      </main>
    </div>
  );
}
