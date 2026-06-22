import axios from 'axios'
import type { NFe, ClientSummary, InternalClient, UploadResponse } from '@/types'

const api = axios.create({ baseURL: '/api/v1', timeout: 30_000 })

export const nfeApi = {
  upload(files: File[]): Promise<UploadResponse> {
    const form = new FormData()
    files.forEach(f => form.append('files', f))
    return api.post<UploadResponse>('/xml/upload', form, {
      headers: { 'Content-Type': 'multipart/form-data' }
    }).then(r => r.data)
  },

  listAll(): Promise<NFe[]> {
    return api.get<{ data: NFe[] }>('/nfe').then(r => r.data.data)
  },

  listUnidentified(): Promise<NFe[]> {
    return api.get<{ data: NFe[] }>('/nfe/unidentified').then(r => r.data.data)
  },

  listQuarantine(): Promise<NFe[]> {
    return api.get<{ data: NFe[] }>('/nfe/quarantine').then(r => r.data.data)
  },

  clientSummary(): Promise<ClientSummary[]> {
    return api.get<{ data: ClientSummary[] }>('/nfe/summary').then(r => r.data.data)
  },

  listClients(): Promise<InternalClient[]> {
    return api.get<{ data: InternalClient[] }>('/clients').then(r => r.data.data)
  }
}
