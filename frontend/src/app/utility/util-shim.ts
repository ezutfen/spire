// Browser-safe replacement for the Node `util` module.
//
// The legacy frontend imported Node's `util` solely to call `util.format(...)`.
// Importing a Node builtin in the browser only worked because the npm `util`
// polyfill was aliased in vite.config.ts; this shim removes that dependency.
//
// `format` implements the printf-style subset of Node's `util.format` that the
// frontend relies on (`%s`, `%d`/`%i`, `%f`, `%j`/`%o`/`%O`, `%%`). It follows
// Node's semantics: extra arguments are appended space-separated, missing
// arguments leave their specifier intact, and `%%` emits a literal percent.

function inspectValue(value: unknown): string {
  if (value === null) {
    return "null"
  }
  if (value === undefined) {
    return "undefined"
  }
  const type = typeof value
  if (type === "string") {
    return value as string
  }
  if (type === "number" || type === "bigint" || type === "boolean") {
    return String(value)
  }
  if (type === "symbol") {
    return (value as symbol).toString()
  }
  if (type === "function") {
    const name = (value as { name?: string }).name || "anonymous"
    return `[Function: ${name}]`
  }
  try {
    return JSON.stringify(value)
  } catch {
    return String(value)
  }
}

export function format(format?: unknown, ...args: unknown[]): string {
  if (typeof format !== "string") {
    return [format, ...args].map(inspectValue).join(" ")
  }

  let i = 0
  let argIndex = 0
  let out = ""

  while (i < format.length) {
    const ch = format[i]

    if (ch !== "%") {
      out += ch
      i += 1
      continue
    }

    const specifier = format[i + 1]

    // Trailing lone "%"
    if (!specifier) {
      out += "%"
      i += 1
      continue
    }

    // Literal percent
    if (specifier === "%") {
      out += "%"
      i += 2
      continue
    }

    i += 2
    const arg = args[argIndex]

    switch (specifier) {
      case "s": {
        argIndex += 1
        out += typeof arg === "string" ? arg : inspectValue(arg)
        break
      }
      case "d":
      case "i": {
        argIndex += 1
        if (typeof arg === "bigint") {
          out += String(arg)
        } else if (typeof arg === "number") {
          out += Number.isFinite(arg) ? Math.trunc(arg).toString() : "NaN"
        } else if (typeof arg === "string") {
          const parsed = parseInt(arg, 10)
          out += Number.isNaN(parsed) ? "NaN" : parsed.toString()
        } else {
          out += "NaN"
        }
        break
      }
      case "f": {
        argIndex += 1
        if (typeof arg === "number") {
          out += Number.isFinite(arg) ? String(arg) : "NaN"
        } else if (typeof arg === "string") {
          const parsed = parseFloat(arg)
          out += Number.isNaN(parsed) ? "NaN" : parsed.toString()
        } else {
          out += "NaN"
        }
        break
      }
      case "j":
      case "o":
      case "O": {
        argIndex += 1
        try {
          out += JSON.stringify(arg)
        } catch {
          out += "[Circular]"
        }
        break
      }
      case "c": {
        // CSS specifier is consumed but ignored outside DOM/console contexts.
        argIndex += 1
        break
      }
      default: {
        // Unknown specifier: leave it untouched, do not consume an argument.
        out += "%" + specifier
        break
      }
    }
  }

  // Extra arguments are appended, space-separated.
  if (argIndex < args.length) {
    const extra = args.slice(argIndex).map(inspectValue).join(" ")
    out += (out.length ? " " : "") + extra
  }

  return out
}

export default { format }
