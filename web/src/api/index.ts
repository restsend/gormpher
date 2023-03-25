import request from './request'

const serverPrefix = (window as any).serverPrefix
// const serverPrefix = ''

export default {
  getObjectNames: () => request.get(`${serverPrefix}/object_names`),
  getObject: (name: string) => request.get(`${serverPrefix}/object/${name}`),
  getObjInfo: (name: string) => request.get(`${serverPrefix}/obj_info/${name}`),
}
