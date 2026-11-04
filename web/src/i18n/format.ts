import { english, type MessageKey } from './messages'
import * as labels from './labels'
export type Locale = 'id' | 'en'
export const validLocale = (value: unknown): value is Locale => value === 'id' || value === 'en'
export function createI18n(locale: Locale) {
  const intlLocale = locale === 'en' ? 'en-US' : 'id-ID'
  const sourceText = (text: string): string => {
    if (locale !== 'en') return text
    if (Object.hasOwn(english, text)) return english[text as MessageKey]
    for (const [prefix, translated] of [
      ['Menjalankan ', 'Running '],
      ['Bukti tersimpan: ', 'Stored evidence: '],
    ]) {
      if (text.startsWith(prefix)) {
        const name = text.slice(prefix.length)
        const label = labels.toolLabel[name]
        return translated + (label ? sourceText(label) : name)
      }
    }
    return text
  }
  const t = (key: MessageKey) => sourceText(key)
  const mapLabels = (values: Record<string, string>) =>
    Object.fromEntries(Object.entries(values).map(([key, value]) => [key, sourceText(value)]))
  const money = new Intl.NumberFormat(intlLocale, {
    style: 'currency',
    currency: 'IDR',
    maximumFractionDigits: 0,
  })
  const number = new Intl.NumberFormat(intlLocale)
  const date = new Intl.DateTimeFormat(intlLocale, {
    timeZone: 'Asia/Bangkok',
    day: '2-digit',
    month: 'short',
    hour: '2-digit',
    minute: '2-digit',
    hourCycle: 'h23',
  })
  return {
    locale,
    t,
    sourceText,
    stateLabel: mapLabels(labels.stateLabel),
    ruleLabel: mapLabels(labels.ruleLabel),
    toolLabel: mapLabels(labels.toolLabel),
    factLabel: mapLabels(labels.factLabel),
    timestamp: (value: string) => date.format(new Date(value)) + ' WIB',
    factValue: (key: string, value: string) => {
      if (key === 'expense_posting_closed') return value === 'true' ? t('Ya') : t('Belum')
      if (key.endsWith('_rupiah')) return money.format(BigInt(value))
      if (key.endsWith('_bps')) return number.format(Number(value) / 100) + '%'
      if (key.endsWith('_minutes')) return number.format(BigInt(value)) + ' ' + t('menit')
      return sourceText(value)
    },
  }
}
