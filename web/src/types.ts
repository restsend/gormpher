export type ActionType = 'filter' | 'search' | 'order' | 'edit'

export interface TableState {
  name: string
  fields: string[]
  types: string[]
  mapping: Record<string, string> // map field -> type
  goMapping: Record<string, string> // map field -> goType
  filters: string[]
  orders: string[]
  searchs: string[]
  edits: string[]
}

export type FilterOp = '=' | '<>' | 'in' | 'not_in' | '>' | '>=' | '<' | '<='
export type OrderOp = 'asc' | 'desc'

export interface Filter {
  name: string
  op: FilterOp
  value: any
}

export interface Order {
  name: string
  op: OrderOp
}
