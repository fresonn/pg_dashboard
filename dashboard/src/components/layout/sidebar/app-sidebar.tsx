import React from 'react'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem
} from '@/components/ui/shadcn/sidebar'
import { sidebarData } from './data'
import { NavGroup } from './nav-group'
import { Header } from './header'
import { DisconnectDialog } from '../disconnect'
import { Separator } from '@/components/ui/shadcn/separator'

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  return (
    <Sidebar collapsible="icon" variant="inset" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <Header />
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        {sidebarData.navGroups.map((props) => (
          <NavGroup key={props.title} {...props} />
        ))}
      </SidebarContent>
      <SidebarFooter className="px-0">
        <Separator />
        <SidebarGroup>
          <SidebarMenu>
            <SidebarMenuItem>
              <DisconnectDialog />
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>
      </SidebarFooter>
    </Sidebar>
  )
}
