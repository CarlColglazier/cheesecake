import axios from 'axios'

let base = process.server ? 'http://backend:8080/' : 'http://localhost:8080/'
if (process.env.ENVIRONMENT === 'production') {
	base = 'https://cheesecake.live/'
}

export default axios.create({
	baseURL: base
})
