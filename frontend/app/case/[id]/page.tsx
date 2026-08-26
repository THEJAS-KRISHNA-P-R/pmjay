import { redirect } from "next/navigation";

/**
 * The case workspace moved to /cases/[id] (see app/cases/[id]/page.tsx)
 * as part of the wider Dashboard/New Case/Settings restructure — /case
 * (singular) is now just the app-area namespace's leftover, /cases
 * (plural) matches the rest of it. This redirect exists so any
 * previously-shared or bookmarked /case/:id link keeps working exactly
 * as before, since a case's URL is the only thing standing between a
 * family and their own case (no login — see ARCHITECTURE.md) and links
 * to it may already be saved in a text message or a browser bookmark.
 */
export default async function LegacyCaseRedirect({
  params,
}: {
  params: Promise<{ id: string }>;
}) {
  const { id } = await params;
  redirect(`/cases/${id}`);
}
