export type Database = {
  id: string
  name: string
  owner: string
  sizePretty: string
  isTemplate: boolean
  allowConnections: boolean
  activeConnections: number
  encoding: string
  collate: string
  ctype: string
  connectionLimit: number
  sizeBytes: number
}
