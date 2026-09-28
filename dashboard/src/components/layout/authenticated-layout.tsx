import { Outlet } from '@tanstack/react-router'
import { AppSidebar } from './sidebar/app-sidebar'
import { SidebarProvider, getSidebarState } from '../ui/shadcn/sidebar'

export function AuthenticatedLayout() {
  return (
    <SidebarProvider defaultOpen={getSidebarState()}>
      <div className="flex h-screen w-full">
        <AppSidebar className="relative w-62" />
        <div className="py-2">
          <div className="shadow-main h-full flex-1 overflow-auto rounded-lg bg-[#111113] px-5 pt-2 pb-7">
            <Outlet />
          </div>
        </div>
      </div>
    </SidebarProvider>
  )
}
