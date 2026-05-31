import axios from 'axios'

const apiClient = axios.create({
    baseURL: '/api',
    withCredentials: true,
})

export const fetchJSON = async (url: string) => {
    const response = await apiClient.get(url)
    return response.data
}

export const postJSON = async (url: string, data: any) => {
    const response = await apiClient.post(url, data)
    return response.data
}

export default apiClient
