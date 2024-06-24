import {createRouter, createWebHashHistory, RouteRecordRaw} from "vue-router";

export const routes: Array<RouteRecordRaw> = [
    // 首页
    {
        path: "/",
        name: "Layout",
        meta: {
            title: "首页",
        },
        component: () => import("@/layout/main.vue"),
        children: [
            {
                path: "",
                alias: "home",
                name: "Home",
                meta: {
                    title: "首页",
                },
                component: () => import("@/views/Home.vue"),
            },
            {
                path: "gateway",
                name: "Gateway",
                meta: {
                    title: "网关管理",
                },
                component: () => import("@/views/Gateway.vue"),
            },
            {
                path: "upstream",
                name: "Upstream",
                meta: {
                    title: "上游管理",
                },
                component: () => import("@/views/Upstream.vue"),
            },
            {
                path: "certificate",
                name: "Certificate",
                meta: {
                    title: "证书管理",
                },
                component: () => import("@/views/Certificate.vue"),
            },
            {
                path: "user",
                name: "User",
                meta: {
                    title: "用户管理",
                },
                component: () => import("@/views/User.vue"),
            },
        ],
    },
];

const router = createRouter({
    history: createWebHashHistory(),
    routes,
});

router.beforeEach((to, _, next) => {
    document.title = (to.meta.title as string) + import.meta.env.VITE_APP_TITLE;
    next();
});

export default router;
