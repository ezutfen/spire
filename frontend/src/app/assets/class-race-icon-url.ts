const classRaceIconUrls = Object.fromEntries(
  Object.entries(
    import.meta.glob('../../assets/img/icons/classes-races/*.png', {
      eager: true,
      import: 'default',
    }),
  ).map(([path, url]) => {
    const fileName = path.split('/').pop()?.replace('.png', '') || ''
    return [fileName, url as string]
  }),
)

export function getClassRaceIconUrl(iconId: number | string | null | undefined): string | null {
  if (iconId === null || typeof iconId === 'undefined' || iconId === '') {
    return null
  }

  return classRaceIconUrls[`item_${iconId}`] || null
}
