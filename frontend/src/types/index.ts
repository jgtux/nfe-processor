export type OperationType = 'purchase' | 'sale' | 'unidentified'
export type ProcessingStatus = 'pending' | 'processed' | 'error'

export interface NFe {
  id: string
  upload_id: string
  access_key: string
  issuer_name: string
  issuer_cnpj: string
  recipient_name: string
  recipient_cnpj: string
  total_amount: number
  issued_at: string
  operation: OperationType
  linked_client?: string
  unidentified_note?: string
  status: ProcessingStatus
  error_msg?: string
  created_at: string
}

export interface ClientSummary {
  client: string
  purchases: number
  sales: number
}

export interface InternalClient {
  id: string
  name: string
  cnpj: string
}

export interface UploadFileResult {
  file: string
  id?: string
  error?: string
}

export interface UploadResponse {
  data: {
    message: string
    results: UploadFileResult[]
  }
}
