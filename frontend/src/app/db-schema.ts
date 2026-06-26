import {SpireApi} from "./api/spire-api";
import { stringFormat } from "@/app/utility/string-format";

export class DbSchema {
  public static async getTableSchema(tableName: string) {
    const r = await SpireApi.v1().get(stringFormat("query/schema/table/%s", tableName))
    if (r.data && r.data.data) {
      return r.data.data
    }

    return {}
  }

  public static async getTableColumns(tableName: string) {
    const schema = await this.getTableSchema(tableName)
    let fields   = <any>[]
    if (schema) {
      schema.forEach((e) => {
        fields.push(e.Column)
      })
    }

    return fields
  }
}
