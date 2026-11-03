import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory } from 'vue-router'
import App from './App.vue'
import HomeView from './views/HomeView.vue'
import FavoritesView from './views/FavoritesView.vue'
import SongView from './views/SongView.vue'
import QuizView from './views/QuizView.vue'
import LessonsView from './views/LessonsView.vue'
import './assets/main.css'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/', name: 'home', component: HomeView },
    { path: '/favorites', name: 'favorites', component: FavoritesView },
    { path: '/lessons', name: 'lessons', component: LessonsView },
    { path: '/lessons/:id', name: 'song', component: SongView, props: true },
    { path: '/quiz', name: 'quiz', component: QuizView },
  ],
})

const app = createApp(App)
app.use(createPinia())
app.use(router)
app.mount('#app')