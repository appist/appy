import { addMessages, init } from "svelte-i18n";
import messages from "@/locales/en.json";

const locale = "en";
addMessages(locale, messages);
init({
  fallbackLocale: "en",
  initialLocale: locale,
});
