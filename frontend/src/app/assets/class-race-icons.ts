const TRANSPARENT_PIXEL =
  'data:image/gif;base64,R0lGODlhAQABAIAAAMLCwgAAACH5BAAAAAAALAAAAAABAAEAAAICRAEAOw=='

const iconModules = import.meta.glob('/src/assets/img/icons/classes-races/*.png', {
  eager: true,
  import: 'default',
}) as Record<string, string>

const iconUrlsById = Object.fromEntries(
  Object.entries(iconModules)
    .map(([path, url]) => {
      const match = path.match(/item_(\d+)\.png$/)
      return match ? [match[1], url] : null
    })
    .filter((entry): entry is [string, string] => entry !== null),
)

export function getClassRaceIconUrl(iconId?: number | string | null) {
  if (iconId === undefined || iconId === null) {
    return TRANSPARENT_PIXEL
  }

  return iconUrlsById[String(iconId)] || TRANSPARENT_PIXEL
}
