import { defineStore } from 'pinia'
import { generateColorVariants, generateColorWithOpacity } from '@/utils/colorUtils'

export interface ThemeColors {
  bgHeaderStart: string
  bgHeaderEnd: string
  elPrimary: string
  elSuccess: string
  elWarning: string
  elDanger: string
  elInfo: string
  primary: string
  secondary: string
  accent: string
  bgMain: string
  bgCard: string
  bgPanel: string
  bgPanelHeader: string
  bgTableHover: string
  bgActive: string
  bgLighter: string
  bgMask: string
  bgOverlay: string
  bgPanelOpaque: string
  bgPanelOpaqueDarker: string
  bgBlank: string
  borderCard: string
  borderMain: string
  borderActive: string
  borderTable: string
  borderLight: string
  textMain: string
  textSecondary: string
  textMuted: string
  textWhite: string
  textPrimary: string
  textButton: string
  alarm: string
  topNavbarActive: string
}

export interface Theme {
  name: string
  displayName: string
  colors: ThemeColors
}

const themes: Record<string, Theme> = {
  light: {
    name: 'light',
    displayName: '明亮',
    colors: {
      bgHeaderStart: '#004999',
      bgHeaderEnd: '#1a7be6',
      elPrimary: '#409eff',
      elSuccess: '#19be6b',
      elWarning: '#ff9900',
      elDanger: '#f56c6c',
      elInfo: '#999',
      primary: '#409EFF',
      secondary: '#409EFF',
      accent: '#409EFF',
      bgMain: '#eee',
      bgCard: 'rgba(229, 229, 229, 0.37)',
      bgPanel: 'rgba(244,244,244,0.62)',
      bgPanelHeader: 'rgba(255, 255, 255, 0.8)',
      bgTableHover: 'rgba(0,89,255,0.13)',
      bgActive: 'rgb(211 233 255)',
      bgPanelOpaque: 'rgba(255, 255, 255, 1)',
      bgPanelOpaqueDarker: 'rgb(235, 235, 235)',
      bgLighter: 'rgba(243,243,243,0.4)',
      bgMask: 'rgba(0, 0, 0, 0.5)',
      bgBlank: 'rgba(229, 229, 229, 0.37)',
      bgOverlay: 'white',
      borderCard: 'rgb(217, 217, 217)',
      borderMain: 'rgba(196, 196, 196, 0.41)',
      borderActive: '#409EFF',
      borderTable: 'rgba(0, 0, 0, 0.1)',
      borderLight: 'rgba(196, 196, 196, 0.2)',
      textMain: '#444',
      textSecondary: '#666666',
      textMuted: '#999999',
      textWhite: '#ffffff',
      textPrimary: '#333333',
      textButton: '#333',
      alarm: '#f56c6c',
      topNavbarActive: '#409EFF',
    },
  },
}

function generateElementPlusColors(colors: ThemeColors) {
  const primaryVariants = generateColorVariants(colors.elPrimary)
  const successVariants = generateColorVariants(colors.elSuccess)
  const warningVariants = generateColorVariants(colors.elWarning)
  const dangerVariants = generateColorVariants(colors.elDanger)
  const infoVariants = generateColorVariants(colors.elInfo)
  return {
    '--el-color-primary': colors.elPrimary,
    '--el-color-primary-light-3': primaryVariants.light3,
    '--el-color-primary-light-5': primaryVariants.light5,
    '--el-color-primary-light-7': primaryVariants.light7,
    '--el-color-primary-light-8': primaryVariants.light8,
    '--el-color-primary-light-9': primaryVariants.light9,
    '--el-color-primary-dark-2': primaryVariants.dark2,
    '--el-color-success': colors.elSuccess,
    '--el-color-success-light-3': successVariants.light3,
    '--el-color-success-light-5': successVariants.light5,
    '--el-color-success-light-7': successVariants.light7,
    '--el-color-success-light-8': successVariants.light8,
    '--el-color-success-light-9': successVariants.light9,
    '--el-color-success-dark-2': successVariants.dark2,
    '--el-color-warning': colors.elWarning,
    '--el-color-warning-light-3': warningVariants.light3,
    '--el-color-warning-light-5': warningVariants.light5,
    '--el-color-warning-light-7': warningVariants.light7,
    '--el-color-warning-light-8': warningVariants.light8,
    '--el-color-warning-light-9': warningVariants.light9,
    '--el-color-warning-dark-2': warningVariants.dark2,
    '--el-color-danger': colors.elDanger,
    '--el-color-danger-light-3': dangerVariants.light3,
    '--el-color-danger-light-5': dangerVariants.light5,
    '--el-color-danger-light-7': dangerVariants.light7,
    '--el-color-danger-light-8': dangerVariants.light8,
    '--el-color-danger-light-9': dangerVariants.light9,
    '--el-color-danger-dark-2': dangerVariants.dark2,
    '--el-color-error': colors.elDanger,
    '--el-color-info': colors.elInfo,
    '--el-color-info-light-3': infoVariants.light3,
    '--el-color-info-light-5': infoVariants.light5,
    '--el-color-info-light-7': infoVariants.light7,
    '--el-color-info-light-8': infoVariants.light8,
    '--el-color-info-light-9': infoVariants.light9,
    '--el-color-info-dark-2': infoVariants.dark2,
  }
}

export const useThemeStore = defineStore('theme', {
  state: () => ({
    currentTheme: 'light' as string,
    themes,
    isLoaded: false,
  }),
  getters: {
    currentThemeData: (state) => state.themes[state.currentTheme],
    availableThemes: (state) => Object.values(state.themes),
    cssVariables: (state) => {
      const theme = state.themes[state.currentTheme]
      if (!theme) return {}
      const primarySelected = theme.colors.secondary
      const primaryHover = generateColorWithOpacity(theme.colors.secondary, 0.6)
      const primarySubtle = generateColorWithOpacity(theme.colors.secondary, 0.1)
      return {
        '--theme-bg-header-start': theme.colors.bgHeaderStart,
        '--theme-bg-header-end': theme.colors.bgHeaderEnd,
        '--theme-primary': theme.colors.primary,
        '--theme-primary-selected': primarySelected,
        '--theme-primary-hover': primaryHover,
        '--theme-primary-subtle': primarySubtle,
        '--theme-secondary': theme.colors.secondary,
        '--theme-accent': theme.colors.accent,
        '--theme-bg-main': theme.colors.bgMain,
        '--theme-bg-card': theme.colors.bgCard,
        '--theme-bg-panel': theme.colors.bgPanel,
        '--theme-bg-panel-header': theme.colors.bgPanelHeader,
        '--theme-bg-table-hover': theme.colors.bgTableHover,
        '--theme-bg-opaque': theme.colors.bgPanelOpaque,
        '--theme-bg-opaque-darker': theme.colors.bgPanelOpaqueDarker,
        '--theme-bg-active': theme.colors.bgActive,
        '--theme-bg-lighter': theme.colors.bgLighter,
        '--theme-border-card': theme.colors.borderCard,
        '--theme-border-main': theme.colors.borderMain,
        '--theme-border-active': theme.colors.borderActive,
        '--theme-border-table': theme.colors.borderTable,
        '--theme-border-light': theme.colors.borderLight,
        '--theme-text-main': theme.colors.textMain,
        '--theme-text-secondary': theme.colors.textSecondary,
        '--theme-text-muted': theme.colors.textMuted,
        '--theme-text-white': theme.colors.textWhite,
        '--theme-alarm': theme.colors.alarm,
        '--theme-top-navbar-active': theme.colors.topNavbarActive,
        '--el-bg-color': theme.colors.bgMain,
        '--el-bg-color-overlay': theme.colors.bgPanelOpaque,
        '--el-button-bg-color': theme.colors.bgPanelOpaque,
        '--el-text-color-regular': theme.colors.textMain,
        '--el-text-color-primary': theme.colors.textPrimary,
        '--el-text-color-secondary': theme.colors.textSecondary,
        '--el-border-color': theme.colors.borderCard,
        '--el-border-color-hover': theme.colors.borderActive,
        '--el-dialog-bg-color': theme.colors.bgPanelOpaque,
        '--el-mask-color': theme.colors.bgMask,
        '--el-card-bg-color': theme.colors.bgPanelOpaque,
        '--el-fill-color-blank': theme.colors.bgBlank,
        '--el-input-text-color': theme.colors.textMain,
        '--el-text-color-placeholder': theme.colors.textMuted,
        '--el-input-focus-border-color': theme.colors.borderActive,
        '--el-table-header-text-color': theme.colors.textMain,
        '--el-table-border-color': theme.colors.borderTable,
        '--el-border-color-light': theme.colors.borderTable,
        '--el-border-color-lighter': theme.colors.borderTable,
        '--el-menu-bg-color': theme.colors.bgMain,
        '--el-menu-text-color': theme.colors.textMain,
        '--el-menu-active-color': theme.colors.textMain,
        '--el-fill-color': theme.colors.bgPanelHeader,
        '--el-fill-color-light': theme.colors.bgPanelHeader,
        '--el-fill-color-lighter': theme.colors.bgLighter,
        '--el-table-row-hover-bg-color': theme.colors.bgPanelHeader,
        '--el-table-header-bg-color': theme.colors.bgPanelHeader,
        '--el-button-text-color': theme.colors.textButton,
        ...generateElementPlusColors(theme.colors),
      }
    },
  },
  actions: {
    setTheme(themeName: string) {
      if (this.themes[themeName]) {
        this.currentTheme = themeName
        this.applyTheme()
        localStorage.setItem('app-theme', this.currentTheme)
      }
    },
    applyTheme() {
      Object.entries(this.cssVariables).forEach(([property, value]) => {
        document.documentElement.style.setProperty(property, String(value))
      })
      this.isLoaded = true
    },
    loadTheme() {
      const savedTheme = localStorage.getItem('app-theme')
      if (savedTheme && this.themes[savedTheme]) {
        this.currentTheme = savedTheme
      }
      this.applyTheme()
    },
  },
})
