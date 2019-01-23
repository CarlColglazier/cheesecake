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

/**
 * Vue filter to round the decimal to the given place.
 * https://gist.github.com/belsrc/672b75d1f89a9a5c192c
 * http://jsfiddle.net/bryan_k/3ova17y9/
 *
 * @param {String} value    The value string.
 * @param {Number} decimals The number of decimal places.
 */
Vue.filter('round', function(value, decimals) {
  if(!value) {
    value = 0;
  }

  if(!decimals) {
    decimals = 0;
  }

  value = Math.round(value * Math.pow(10, decimals)) / Math.pow(10, decimals);
  return value;
});

/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  components: { App },
  template: '<App/>'
})
