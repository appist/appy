import "@testing-library/jest-dom/extend-expect";
import { render } from "@testing-library/svelte";

import NotFound from "@/pages/NotFound.svelte";

test("renders the not found message", () => {
  const { getByText } = render(NotFound);

  expect(getByText("Oops! The page you are looking for doesn't exist.")).toBeTruthy();
});
