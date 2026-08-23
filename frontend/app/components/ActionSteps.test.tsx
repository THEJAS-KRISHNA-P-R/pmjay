import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ActionSteps } from "./ActionSteps";

describe("ActionSteps", () => {
  it("renders nothing at all for an empty steps array, rather than an empty section", () => {
    // app/case/[id]/page.tsx already guards this with
    // `action_steps.length > 0` before rendering the component at all,
    // but the component has its own independent guard too (defense in
    // depth, same reasoning as CareFirstBanner not trusting the backend
    // alone) -- this test is what actually exercises that second guard.
    const { container } = render(<ActionSteps steps={[]} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("renders every step's text", () => {
    render(<ActionSteps steps={["Ask for the discharge summary", "Call the helpline"]} />);
    expect(screen.getByText("Ask for the discharge summary")).toBeInTheDocument();
    expect(screen.getByText("Call the helpline")).toBeInTheDocument();
  });

  it("numbers steps in the order given, starting at 1", () => {
    render(<ActionSteps steps={["First thing", "Second thing", "Third thing"]} />);
    const list = screen.getByRole("list");
    const items = list.querySelectorAll("li");
    expect(items).toHaveLength(3);
    // The number badge is a separate aria-hidden span from the step
    // text (see source) -- checking the rendered order of the <li>
    // elements themselves is what actually verifies steps aren't
    // silently reordered, rather than trusting the numbering alone.
    expect(items[0]).toHaveTextContent("First thing");
    expect(items[1]).toHaveTextContent("Second thing");
    expect(items[2]).toHaveTextContent("Third thing");
    for (const [i, item] of Array.from(items).entries()) {
      expect(item).toHaveTextContent(String(i + 1));
    }
  });

  it("marks the number badge aria-hidden, since list order already conveys sequence to assistive tech", () => {
    render(<ActionSteps steps={["Only step"]} />);
    const badge = screen.getByText("1");
    expect(badge).toHaveAttribute("aria-hidden", "true");
  });
});
