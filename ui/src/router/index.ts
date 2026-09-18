import { createRouter, createWebHistory } from "vue-router";

import HomeView from "../views/HomeView.vue";
import LearnView from "../views/LearnView.vue";
import CreateView from "../views/CreateView.vue";

const routes = [
  { path: "/", name: "home", component: HomeView },
  { path: "/learn/:id", name: "learn", component: LearnView },
  { path: "/create", name: "create", component: CreateView },
];

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
});

export default router;
