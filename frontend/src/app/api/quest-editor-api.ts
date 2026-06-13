import {SpireApi} from './spire-api'

export class QuestEditorApi {
  static async getCapabilities() {
    const r = await SpireApi.v1().get('/eqemuserver/quests/capabilities')
    return r.data && r.data.data ? r.data.data : {}
  }

  static async getTree() {
    const r = await SpireApi.v1().get('/eqemuserver/quests/tree')
    return r.data && r.data.data ? r.data.data : []
  }

  static async getFile(path) {
    const r = await SpireApi.v1().get('/eqemuserver/quests/file', {params: {path}})
    return r.data && r.data.data ? r.data.data : null
  }

  static async saveFile(path, contents) {
    return SpireApi.v1().put('/eqemuserver/quests/file', {path, contents})
  }

  static async createFile(relativePath, content = '') {
    return SpireApi.v1().post('/eqemuserver/quests/file/create', {
      relative_path: relativePath,
      content: content,
    })
  }

  static async createFolder(relativePath) {
    return SpireApi.v1().post('/eqemuserver/quests/folder/create', {
      relative_path: relativePath,
    })
  }

  static async movePath(oldPath, newPath) {
    return SpireApi.v1().post('/eqemuserver/quests/path/move', {
      old_path: oldPath,
      new_path: newPath,
    })
  }

  static async deletePath(path) {
    return SpireApi.v1().delete('/eqemuserver/quests/path', {params: {path}})
  }

  static async formatFile(path, contents) {
    const r = await SpireApi.v1().post('/eqemuserver/quests/file/format', {path, contents})
    return r.data && r.data.data ? r.data.data : null
  }

  static async validateFile(path, contents) {
    const r = await SpireApi.v1().post('/eqemuserver/quests/file/validate', {path, contents})
    return r.data && r.data.data ? r.data.data : null
  }

  static async reloadFile(path) {
    return SpireApi.v1().post('/eqemuserver/quests/file/reload', {path})
  }
}
