export function hexToRgb(hex: string): { r: number; g: number; b: number } | null {
  const cleanHex = hex.replace('#', '')
  if (cleanHex.length === 3) {
    return {
      r: parseInt(cleanHex[0] + cleanHex[0], 16),
      g: parseInt(cleanHex[1] + cleanHex[1], 16),
      b: parseInt(cleanHex[2] + cleanHex[2], 16),
    }
  }
  if (cleanHex.length === 6) {
    return {
      r: parseInt(cleanHex.substring(0, 2), 16),
      g: parseInt(cleanHex.substring(2, 4), 16),
      b: parseInt(cleanHex.substring(4, 6), 16),
    }
  }
  return null
}

export function generateLightVariant(baseColor: string, lightness: number): string {
  const rgb = hexToRgb(baseColor)
  if (!rgb) return baseColor
  let mixRatio = 0.5
  switch (lightness) {
    case 3:
      mixRatio = 0.6
      break
    case 5:
      mixRatio = 0.4
      break
    case 7:
      mixRatio = 0.2
      break
    case 8:
      mixRatio = 0.1
      break
    case 9:
      mixRatio = 0.05
      break
    default:
      mixRatio = 0.5
  }
  const r = Math.round(rgb.r * mixRatio + 255 * (1 - mixRatio))
  const g = Math.round(rgb.g * mixRatio + 255 * (1 - mixRatio))
  const b = Math.round(rgb.b * mixRatio + 255 * (1 - mixRatio))
  return `rgb(${r}, ${g}, ${b})`
}

export function generateDarkVariant(baseColor: string, darkness: number): string {
  const rgb = hexToRgb(baseColor)
  if (!rgb) return baseColor
  const mixRatio = darkness === 2 ? 0.8 : 0.8
  return `rgb(${Math.round(rgb.r * mixRatio)}, ${Math.round(rgb.g * mixRatio)}, ${Math.round(rgb.b * mixRatio)})`
}

export function generateColorWithOpacity(baseColor: string, opacity: number): string {
  const rgb = hexToRgb(baseColor)
  if (!rgb) return baseColor
  return `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, ${opacity})`
}

export function generateColorVariants(baseColor: string) {
  return {
    base: baseColor,
    light3: generateLightVariant(baseColor, 3),
    light5: generateLightVariant(baseColor, 5),
    light7: generateLightVariant(baseColor, 7),
    light8: generateLightVariant(baseColor, 8),
    light9: generateLightVariant(baseColor, 9),
    dark2: generateDarkVariant(baseColor, 2),
  }
}
