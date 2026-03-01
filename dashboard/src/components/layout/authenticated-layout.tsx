import { Outlet } from '@tanstack/react-router'
import { AppSidebar } from './sidebar/app-sidebar'
import { SidebarProvider } from '../ui/shadcn/sidebar'

export function AuthenticatedLayout() {
  return (
    <SidebarProvider defaultOpen>
      <div className="flex h-screen">
        <AppSidebar className="relative w-72" />
        <main className="flex-1 overflow-auto px-2 py-4">
          <Outlet />
        </main>
      </div>
    </SidebarProvider>
  )
}
