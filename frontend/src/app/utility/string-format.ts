export function stringFormat(template: string, ...values: unknown[]): string {
  let valueIndex = 0

  const formatted = String(template).replace(/%[s%]/g, (token) => {
    if (token === '%%') {
      return '%'
    }

    if (valueIndex >= values.length) {
      return token
    }

    return String(values[valueIndex++])
  })

  if (valueIndex >= values.length) {
    return formatted
  }

  const extras = values.slice(valueIndex).map((value) => String(value)).join(' ')
  return extras ? `${formatted} ${extras}` : formatted
}
