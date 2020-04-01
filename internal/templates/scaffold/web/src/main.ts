import initI18n from "@/initI18n";
import "@/initServiceWorker";
import App from "@/components/App.svelte";

const replaceContents = (node: HTMLElement | null): HTMLElement | null => {
  if (node) node.innerHTML = "";

  return node;
};

const initApp = async () => {
  await initI18n();

  new App({
    target: replaceContents(document.getElementById("app")) || new HTMLElement(),
  });
};

initApp();
