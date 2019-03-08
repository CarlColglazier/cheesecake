import Vue from 'vue'
import Router from 'vue-router'
import Home from '@/components/Home'
import Event from '@/components/Event'
import Events from '@/components/Events'
import Rankings from '@/components/Rankings'

Vue.use(Router)

export default new Router({
  mode: 'history',
  routes: [
    {
      path: '/',
      name: 'Home',
      component: Home
    },
    {
      path: '/events',
      name: 'Events',
      component: Events
    },
    {
      path: '/event/:key',
      name: 'Event',
      component: Event
    },
    {
      path: '/rankings',
      name: 'Rankings',
      component: Rankings
    }
  ]
})
