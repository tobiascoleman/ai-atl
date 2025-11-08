import apiClient from './client'
import { ChatMessage } from '@/types/api'

export const chatbotAPI = {
  ask: async (question: string): Promise<ChatMessage> => {
    const { data } = await apiClient.post<ChatMessage>('/chatbot/ask', {
      question,
    })
    return data
  },

  getHistory: async () => {
    const { data } = await apiClient.get('/chatbot/history')
    return data
  },
}

