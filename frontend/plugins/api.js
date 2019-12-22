import Vue from 'vue'

let url = (process.server) ? 'http://backend:8080' : 'http://localhost:8080';
if (process.env.ENVIRONMENT === 'production') {
    url = 'https://cheesecake.live/api/';
}

export default ({ app }, inject) => {
    app.fetch = string => fetch(`${url}${string}`)
}
