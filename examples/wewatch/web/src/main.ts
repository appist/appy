import Vue, { CreateElement } from "vue";

import App from "@/App.vue";
import "@/registerServiceWorker";
import router from "@/router";
import store from "@/store";
import i18n, { loadLanguageAsync } from "@/i18n";

Vue.config.productionTip = false;

function renderApp() {
  new Vue({
    i18n,
    render: (h: CreateElement) => h(App),
    router,
    store
  }).$mount("#app");
}

loadLanguageAsync(window.navigator.language)
  .then(() => renderApp())
  .catch(() => renderApp());

if (module.hot) {
  module.hot.accept(
    ["@/locales/en", "@/locales/zh-CN", "@/locales/zh-TW"],
    () => {
      i18n.setLocaleMessage("en", require("@/locales/en").default);
      i18n.setLocaleMessage("zh-CN", require("@/locales/zh-CN").default);
      i18n.setLocaleMessage("zh-TW", require("@/locales/zh-TW").default);
    }
  );
}
