import Vue from "vue";
import VueI18n from "vue-i18n";
import enMessages from "@/locales/en";

Vue.use(VueI18n);

const loadedLocales = ["en"];
const i18n = new VueI18n({
  locale: "en",
  fallbackLocale: "en",
  messages: { en: enMessages }
});

function setI18nLanguage(locale: string) {
  i18n.locale = locale;
  (document.querySelector("html") as HTMLElement).setAttribute("lang", locale);
  return locale;
}

export function loadLanguageAsync(locale: string) {
  if (i18n.locale !== locale) {
    if (!loadedLocales.includes(locale)) {
      return import(
        /* webpackChunkName: "locale-[request]" */ `@/locales/${locale}`
      ).then(msgs => {
        i18n.setLocaleMessage(locale, msgs.default);
        loadedLocales.push(locale);
        return setI18nLanguage(locale);
      });
    }

    return Promise.resolve(setI18nLanguage(locale));
  }

  return Promise.resolve(locale);
}

export default i18n;
