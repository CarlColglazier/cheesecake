import axios from 'axios'

export default axios.create({
	baseURL: process.server ? 'http://backend:8080/' : 'http://localhost:8080/'
})
