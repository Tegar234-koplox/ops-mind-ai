import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import { createI18n, validLocale, type Locale } from './format'
const storageKey = 'opsmind.locale.v1'
const LanguageContext = createContext<{
  locale: Locale
  setLocale: (value: Locale) => void
} | null>(null)
function initialLocale(): Locale {
  try {
    const value = localStorage.getItem(storageKey)
    return validLocale(value) ? value : 'id'
  } catch {
    return 'id'
  }
}
export function LanguageProvider({ children }: { children: ReactNode }) {
  const [locale, setLanguage] = useState<Locale>(initialLocale)
  useEffect(() => {
    document.documentElement.lang = locale
    document.title =
      locale === 'en' ? 'OpsMind AI — Investigation Workspace' : 'OpsMind AI — Ruang Investigasi'
    const description = document.querySelector('meta[name="description"]')
    description?.setAttribute(
      'content',
      locale === 'en'
        ? 'OpsMind AI — evidence-based manufacturing operations investigations.'
        : 'OpsMind AI — investigasi operasional manufaktur berbasis bukti.',
    )
  }, [locale])
  const value = useMemo(
    () => ({
      locale,
      setLocale: (next: Locale) => {
        setLanguage(next)
        try {
          localStorage.setItem(storageKey, next)
        } catch {
          /* Language selection remains available when browser storage is disabled. */
        }
      },
    }),
    [locale],
  )
  return <LanguageContext value={value}>{children}</LanguageContext>
}
export function useI18n() {
  const context = useContext(LanguageContext)
  if (!context) throw new Error('LanguageProvider is required')
  const format = useMemo(() => createI18n(context.locale), [context.locale])
  return { ...format, setLocale: context.setLocale }
}
export function LanguagePicker() {
  const { locale, setLocale, t } = useI18n()
  return (
    <label className="language-picker">
      <span>{t('Bahasa aplikasi')}</span>
      <select
        data-testid="language-picker"
        value={locale}
        onChange={(event) => {
          if (validLocale(event.target.value)) setLocale(event.target.value)
        }}
      >
        <option value="id" lang="id">
          Indonesia
        </option>
        <option value="en" lang="en">
          English
        </option>
      </select>
    </label>
  )
}
export function ReportLanguage({ language }: { language?: string }) {
  const { t } = useI18n()
  return (
    <p className="report-language">
      {t('Bahasa laporan')}:{' '}
      <span lang={language === 'en' ? 'en' : 'id'}>
        {language === 'en' ? 'English' : 'Indonesia'}
      </span>
      . {t('Laporan tersimpan ditampilkan dalam bahasa aslinya.')}
    </p>
  )
}
