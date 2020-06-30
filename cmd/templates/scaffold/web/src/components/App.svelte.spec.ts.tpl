import "@testing-library/jest-dom/extend-expect";
import { render, queryByAttribute } from "@testing-library/svelte";

import App from "@/components/App.svelte";

test("renders nothing", () => {
  const getById = queryByAttribute.bind(null, "id");
  const dom = render(App);

  expect(getById(dom.container, "app")).toBeNull();
});
