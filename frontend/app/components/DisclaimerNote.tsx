import { IconInfo } from "./icons";

/**
 * Renders CaseResponse.disclaimer — sent by the backend on every case,
 * unconditionally (see response/types.go's Disclaimer doc comment: the
 * same structural guarantee mechanism as care_first_message). The PDF
 * has always shown this, directly below its own care-first banner; this
 * component gives the on-screen page the same placement so the two
 * never disagree about what a family was told.
 */
export function DisclaimerNote({ text }: { text: string }) {
  return (
    <div className="flex items-start gap-2.5 px-1 text-xs sm:text-sm leading-relaxed text-sand-500">
      <IconInfo className="h-3.5 w-3.5 shrink-0 mt-0.5" aria-hidden="true" />
      <p>{text}</p>
    </div>
  );
}
