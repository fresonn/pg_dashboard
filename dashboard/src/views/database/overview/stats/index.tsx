import { StaticStatsGrid } from '../ui/static-stats-grid'
import { Widget } from '../ui/widget'
import { Database, Gauge, Layers, Sheet, UsersRound } from 'lucide-react'
import { BufferCacheWidget, BufferCacheWidgetSkeleton } from '@/components/metrics/buffer-cache'
import { SizeWidget, SizeWidgetSkeleton } from './static/size'
import { TablesWidget, TablesWidgetSkeleton } from './static/tables'
import { IndexesWidget, IndexesWidgetSkeleton } from './static/indexes'
import { ConnectionsWidget, ConnectionsWidgetSkeleton } from './сonnections-widget'

export function Stats({ databaseId }: { databaseId: string }) {
  return (
    <div>
      <StaticStatsGrid>
        <Widget
          title="Size"
          icon={<Database />}
          className="col-span-6 row-span-6 lg:col-span-3"
          skeleton={<SizeWidgetSkeleton />}
        >
          <SizeWidget databaseId={databaseId} />
        </Widget>
        <Widget
          title="Tables"
          icon={<Sheet />}
          className="col-span-6 row-span-6 lg:col-span-3"
          skeleton={<TablesWidgetSkeleton />}
        >
          <TablesWidget databaseId={databaseId} />
        </Widget>
        <Widget
          title="Indexes"
          icon={<Layers />}
          className="col-span-6 row-span-6 lg:col-span-3"
          skeleton={<IndexesWidgetSkeleton />}
        >
          <IndexesWidget databaseId={databaseId} />
        </Widget>
        <Widget
          title="Connections"
          icon={<UsersRound />}
          className="col-span-6 row-span-6 lg:col-span-3"
          skeleton={<ConnectionsWidgetSkeleton />}
        >
          <ConnectionsWidget />
        </Widget>
        <Widget
          title="Buffer cache hit rate"
          icon={<Gauge />}
          className="col-span-6 row-span-9 lg:col-span-3"
          skeleton={<BufferCacheWidgetSkeleton />}
        >
          <BufferCacheWidget scope="database" />
        </Widget>
      </StaticStatsGrid>
    </div>
  )
}
