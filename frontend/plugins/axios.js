import axios from 'axios'

let base = 'http://localhost:8080/'
if (process.server) {
	base = 'http://backend:8080/'
} else if (process.env.ENVIRONMENT === 'production') {
	base = 'https://cheesecake.live/api/'
}

export default axios.create({
	baseURL: base
})
