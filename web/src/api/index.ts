import request from './request'

export default {
  getObjectNames: () => request.get('api/object_names'),
  getObject: (name: string) => request.get(`api/object/${name}`),
  getObjInfo: (name: string) => request.get(`api/obj_info/${name}`),
}
