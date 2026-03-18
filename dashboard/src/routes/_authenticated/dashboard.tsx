import { Header } from '@/components/layout/header/header'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_authenticated/dashboard')({
  component: RouteComponent
})

function RouteComponent() {
  return (
    <div>
      <Header title="Dashboard" />
      <main className="w-[1100px]">
        <h1>Hello "/_authenticated/dashboard"!</h1>
      </main>
    </div>
  )
}
