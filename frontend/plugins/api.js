import Vue from 'vue'

let url;
if (process.server) {
    url = 'http://backend:8080'
} else if (process.env.NODE_ENV === 'production') {
    url = 'https://cheesecake.live/api/';
} else {
    url = 'http://localhost:8080';
}

export default ({ app }, inject) => {
    app.fetch = string => fetch(`${url}${string}`)
}
