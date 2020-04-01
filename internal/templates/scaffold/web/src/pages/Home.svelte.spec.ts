import "@testing-library/jest-dom/extend-expect";
import { render } from "@testing-library/svelte";

import Home from "@/pages/Home.svelte";

test("renders the welcome message", () => {
  const { getByText } = render(Home);

  expect(getByText("An opinionated productive web framework that helps scaling business easier.")).toBeTruthy();
});
