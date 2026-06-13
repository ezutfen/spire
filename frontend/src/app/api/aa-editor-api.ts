import {SpireApi} from './spire-api'

export class AaEditorApi {
  static async listAbilities(params = {}) {
    const r = await SpireApi.v1().get('/aa_editor/abilities', {params})
    return r.data && r.data.data ? r.data.data : {total: 0, items: []}
  }

  static async getAbility(id) {
    const r = await SpireApi.v1().get('/aa_editor/abilities/' + id)
    return r.data && r.data.data ? r.data.data : null
  }

  static async getMetadata() {
    const r = await SpireApi.v1().get('/aa_editor/metadata')
    return r.data && r.data.data ? r.data.data : {}
  }

  static async createAbility(payload) {
    const r = await SpireApi.v1().put('/aa_editor/abilities', payload)
    return r.data && r.data.data ? r.data.data : null
  }

  static async duplicateAbility(id, opts = {}) {
    const r = await SpireApi.v1().post('/aa_editor/abilities/' + id + '/duplicate', opts)
    return r.data && r.data.data ? r.data.data : null
  }

  static async saveAbility(id, payload) {
    const r = await SpireApi.v1().patch('/aa_editor/abilities/' + id, payload)
    return r.data && r.data.data ? r.data.data : null
  }

  static async deleteAbility(id) {
    return SpireApi.v1().delete('/aa_editor/abilities/' + id)
  }

  static async previewRank(payload) {
    const r = await SpireApi.v1().post('/aa_editor/ranks/preview', payload)
    return r.data && r.data.data ? r.data.data : null
  }
}
