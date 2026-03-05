import { ThemeToggle } from '@/components/ui/theme-toggle'
import { Typography } from '@/components/ui/typography'
import { VersionWidget, VersionWidgetSkeleton } from './widgets/version'
import { Widget } from './common'
import type { ReactNode } from 'react'
import { UptimeWidget, UptimeWidgetSkeleton } from './widgets/uptime'
import { ClusterSettings, ClusterSettingsSkeleton } from './widgets/cluster-settings'
import { SidebarTrigger } from '@/components/ui/shadcn/sidebar'
import { AvailableDatabases } from './databases'

export function DashboardGrid({ children }: { children: ReactNode }) {
  return (
    <div className="grid grid-cols-[repeat(12,minmax(0,220px))] grid-rows-[repeat(12,15px)] gap-3">
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
            className="col-span-3 row-span-4"
            // withBackground={false}
            skeleton={<VersionWidgetSkeleton />}
          >
            <VersionWidget />
          </Widget>
          <Widget
            title="Uptime"
            className="col-span-3 row-span-4"
            // withBackground={false}
            skeleton={<UptimeWidgetSkeleton />}
          >
            <UptimeWidget />
          </Widget>
          <Widget
            title="Common settings"
            className="col-span-6 row-span-12"
            skeleton={<ClusterSettingsSkeleton />}
          >
            <ClusterSettings />
          </Widget>
        </DashboardGrid>
        <div className="w-full pt-10">
          <Typography variant="h2">Available databases</Typography>
          <div className="pt-4">
            <AvailableDatabases />
          </div>
        </div>
      </div>
    </div>
  )
}
