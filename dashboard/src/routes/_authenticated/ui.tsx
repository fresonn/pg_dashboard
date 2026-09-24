import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { AlarmClockCheck, Database } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Header } from '@/components/layout/header/header'
import { RoleAttributeFlag } from '@/views/overview/roles/ui/attribute-flag'
import { roleAttributesEnum } from '@/lib/api/gen'

export const Route = createFileRoute('/_authenticated/ui')({
  component: RouteComponent
})

export const roleFlag = {
  superuser: 'superuser',
  login: 'login',
  createRole: 'create_role',
  createDb: 'create_db',
  replication: 'replication'
} as const

function RouteComponent() {
  const [loading, setLoading] = useState(false)

  const handleClick = () => {
    setLoading((prev) => !prev)
  }

  return (
    <div>
      <Header title="UI samples" />

      <Button>Default</Button>
      <Button variant="destructive">Click me!</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="link">Click me!</Button>
      <Button variant="light">Light</Button>
      <div className="my-10">
        <Button loading={loading} onClick={handleClick}>
          <Database />
          Loading
        </Button>
        <Button loading={loading} onClick={handleClick} variant="secondary">
          <AlarmClockCheck />
          Loading
        </Button>
        <Button loading={loading} variant="light">
          <Database />
          Light
        </Button>

        <div className="max-w-42">
          <Input placeholder="Port" type="number" />
          <Input disabled placeholder="Port" type="number" />
        </div>
      </div>

      <button
        type="button"
        className="ease-power3-out focus-visible:ring-ring/60 bg-secondary text-secondary-foreground hover:bg-muted group data-[state=open]:bg-muted inline-flex h-[30px] shrink-0 cursor-pointer items-center justify-center gap-0 overflow-hidden rounded-full p-0 text-[12px] font-medium whitespace-nowrap shadow-[0px_0px_0px_1px_rgba(0,0,0,0.4),inset_0px_1px_0px_0px_rgba(255,255,255,0.1),inset_0px_0px_0px_1px_rgba(255,255,255,0.06)] transition-[background-color,color,box-shadow] duration-150 outline-none select-none focus-visible:ring-2 disabled:pointer-events-none disabled:opacity-50"
        id="radix-_R_955flb_"
        aria-haspopup="menu"
        aria-expanded="false"
        data-state="closed"
        data-slot="dropdown-menu-trigger"
      >
        <span className="text-subtle px-[9px] font-normal">Last Activity</span>
        <span aria-hidden="true" className="h-full w-px bg-white/8"></span>
        <span className="flex items-center gap-1.5 px-[9px]">
          90 Days
          <svg
            xmlns="http://www.w3.org/2000/svg"
            width="1em"
            height="1em"
            fill="none"
            viewBox="0 0 12 12"
            aria-hidden="true"
            className="ease-power3-out size-3 text-[#898b8d] transition-transform duration-200 group-data-[state=open]:rotate-180"
          >
            <path
              stroke="currentColor"
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="1.4"
              d="m3 4.5 3 3 3-3"
            ></path>
          </svg>
        </span>
      </button>

      <div className="mb-10">
        <h1>Role attrs</h1>
        {Object.values(roleAttributesEnum).map((attr) => (
          <RoleAttributeFlag key={attr} attribute={attr} />
        ))}
      </div>

      <div>
        <p className="font-code">
          Lorem, ipsum dolor sit amet consectetur adipisicing elit. Vero consequuntur voluptatibus
          laudantium velit nam deserunt quis perspiciatis quaerat voluptas voluptatum labore culpa
          ex nulla dignissimos incidunt, dolor architecto, praesentium ipsam, molestiae nesciunt
          minima quasi ut sit tenetur! Tenetur maxime enim explicabo iure quis aperiam deleniti ab?
          Natus cum a, animi reprehenderit ipsa est magnam numquam mollitia aliquam placeat officia,
          inventore quam alias nam incidunt. Reiciendis nostrum ad alias laborum dolores. Et unde
          corrupti non ratione nulla aperiam eum. Aut quidem, nihil ut nemo temporibus laboriosam
          corporis? Aliquid exercitationem neque velit magni necessitatibus nemo placeat ullam
          minima. Dignissimos aspernatur quis sit.
        </p>
      </div>
    </div>
  )
}
