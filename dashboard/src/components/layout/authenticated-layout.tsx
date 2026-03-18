import { Outlet } from '@tanstack/react-router'
import { AppSidebar } from './sidebar/app-sidebar'
import { SidebarProvider, getSidebarState } from '../ui/shadcn/sidebar'

export function AuthenticatedLayout() {
  return (
    <SidebarProvider defaultOpen={getSidebarState()}>
      <div className="flex h-screen">
        <AppSidebar className="relative w-72" />
        <div className="flex-1 overflow-auto p-3 pt-2">
          <Outlet />
        </div>
      </div>
    </SidebarProvider>
  )
}
