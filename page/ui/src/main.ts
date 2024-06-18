import {createApp} from "vue";
import "./style.css";
import App from "./App.vue";
import router from "@/router";
import store from "@/store";
import "uno.css";
import "@arco-design/web-vue/dist/arco.css";

createApp(App).use(store).use(router).mount("#app");
