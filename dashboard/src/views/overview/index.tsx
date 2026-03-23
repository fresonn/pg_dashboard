import { Header } from '@/components/layout/header/header'
import { Typography } from '@/components/ui/typography'
import { VersionWidget, VersionWidgetSkeleton } from './widgets/version'
import { Widget } from './common'
import type { ReactNode } from 'react'
import { UptimeWidget, UptimeWidgetSkeleton } from './widgets/uptime'
import { ClusterSettings, ClusterSettingsSkeleton } from './widgets/cluster-settings'
import { AvailableDatabases } from './databases'
import { RolesWidget, RolesWidgetSkeleton } from './roles'

export function DashboardGrid({ children }: { children: ReactNode }) {
  return (
    <div className="grid grid-cols-[repeat(12,minmax(0,220px))] grid-rows-[repeat(12,15px)] gap-3">
      {children}
    </div>
  )
}

export function ClusterOverview() {
  return (
    <div>
      <Header title="Cluster Overview" />
      <main>
        <div>
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
            <Widget
              title="Roles & Permissions"
              className="col-span-6 row-span-8"
              skeleton={<RolesWidgetSkeleton />}
            >
              <RolesWidget />
            </Widget>
          </DashboardGrid>
          <div className="w-full pt-10">
            <Typography variant="h2">Available databases</Typography>
            <div className="pt-4">
              <AvailableDatabases />
            </div>
          </div>
        </div>
      </main>
    </div>
  )
}
