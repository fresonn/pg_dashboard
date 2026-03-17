import { useCallback, useState, useEffect } from 'react'

export type CopiedValue = string | null
export type CopyFn = (value: string) => Promise<void>

export const useCopyToClipboard = (): [CopiedValue, CopyFn] => {
  const [copiedText, setCopiedText] = useState<CopiedValue>(null)

  useEffect(() => {
    const t = setTimeout(() => {
      if (copiedText) setCopiedText(null)
    }, 2000)

    return () => {
      clearTimeout(t)
    }
  }, [copiedText])

  const copy: CopyFn = useCallback(async (text) => {
    if (!navigator.clipboard) return Promise.reject(new Error('clipboard not supported'))

    try {
      await navigator.clipboard.writeText(text)
      setCopiedText(text)
    } catch (error) {
      setCopiedText(null)
      return Promise.reject(error)
    }
  }, [])

  return [copiedText, copy]
}
