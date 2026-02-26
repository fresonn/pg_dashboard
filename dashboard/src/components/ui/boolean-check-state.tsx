import { Check, X } from 'lucide-react'

export function BooleanCheckState({ check }: { check: boolean }) {
  return <div>{check ? <Check /> : <X />}</div>
}
