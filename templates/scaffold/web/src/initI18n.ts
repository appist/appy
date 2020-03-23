import { addMessages, init, getLocaleFromNavigator } from "svelte-i18n";

export default async () => {
  let locale = getLocaleFromNavigator();
  if (process.env.AVAILABLE_LOCALES && process.env.AVAILABLE_LOCALES.indexOf(locale) < 0) {
    locale = locale.split("-")[0];
  }

  const messages = await import(`@/locales/${locale}.json`);
  addMessages(locale, messages.default);

  init({
    fallbackLocale: "en",
    initialLocale: locale,
  });
};
