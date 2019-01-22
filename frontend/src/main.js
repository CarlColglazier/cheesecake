// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'
import VueResource from 'vue-resource'
import App from './App'
import router from './router'

Vue.config.productionTip = false

// https://stackoverflow.com/questions/43076464/support-2-base-urls-in-vue
var isProduction = process.env.NODE_ENV === 'production'
var rootUrl = (isProduction) ? 'https://cheesecake.live/api/' : 'http://localhost:5000/api/'
Vue.use(VueResource)
Vue.http.options.root = rootUrl

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  components: { App },
  template: '<App/>'
})
