import { ThemeToggle } from '@/components/ui/theme-toggle'
import { Typography } from '@/components/ui/typography'
import { VersionWidget, VersionWidgetSkeleton } from './widgets/version'
import { Widget } from './common'
import type { ReactNode } from 'react'
import { UptimeWidget, UptimeWidgetSkeleton } from './widgets/uptime'
import { ClusterSettings, ClusterSettingsSkeleton } from './widgets/cluster-settings'
import { SidebarTrigger } from '@/components/ui/shadcn/sidebar'
import { DatabasesTable } from './databases/table'

export function DashboardGrid({ children }: { children: ReactNode }) {
  return (
    <div className="grid grid-cols-[repeat(12,minmax(0,220px))] grid-rows-[repeat(6,50px)] gap-3">
      {children}
    </div>
  )
}

export function ClusterOverview() {
  return (
    <div className="px-1 py-5">
      <Typography variant="h1" as="h1">
        Overview
        <SidebarTrigger variant="outline" className="scale-125 sm:scale-100" />
      </Typography>
      <ThemeToggle />

      <div className="mt-2">
        <DashboardGrid>
          <Widget
            title="Version"
            className="col-span-3 row-span-2"
            skeleton={<VersionWidgetSkeleton />}
          >
            <VersionWidget />
          </Widget>
          <Widget
            title="Uptime"
            className="col-span-3 row-span-2"
            skeleton={<UptimeWidgetSkeleton />}
          >
            <UptimeWidget />
          </Widget>
          {/* <div className="animate-in fade-in in-from-top col-span-6 row-span-3 duration-300 dark:bg-neutral-800"></div> */}
          <Widget
            title="Common cluster settings"
            className="col-span-6 row-span-5"
            withBackground={true}
            skeleton={<ClusterSettingsSkeleton />}
          >
            <ClusterSettings />
          </Widget>
        </DashboardGrid>
        <DatabasesTable />
      </div>
    </div>
  )
}
