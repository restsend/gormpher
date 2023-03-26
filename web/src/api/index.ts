import request from './request'
import type { QueryParams } from '@/views/table/useTable'

export const serverPrefix = (window as any).serverPrefix

export default {
  getObjectNames: () => request.get(`${serverPrefix}/object_names`),
  getObject: (name: string) => request.get(`${serverPrefix}/object/${name}`),

  handleAdd: (name: string, item: any) => request.put(`${serverPrefix}/${name}`, item),
  handleEdit: (name: string, item: any) => request.patch(`${serverPrefix}/${name}/${item.id}`, item),
  handleDelete: (name: string, id: string | number) => request.delete(`${serverPrefix}/${name}/${id}`),
  handleQuery: (name: string, params: QueryParams) => request.post(`${serverPrefix}/${name}`, params),
  handleBatch: (name: string, ids: string[]) => request.delete(`${serverPrefix}/${name}`, ids),
}
