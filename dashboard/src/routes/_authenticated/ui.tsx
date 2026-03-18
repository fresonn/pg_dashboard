import { createFileRoute } from '@tanstack/react-router'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { AlarmClockCheck, Database } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Header } from '@/components/layout/header/header'

export const Route = createFileRoute('/_authenticated/ui')({
  component: RouteComponent
})

function RouteComponent() {
  const [loading, setLoading] = useState(false)

  const hanleClick = () => {
    setLoading((prev) => !prev)
  }

  return (
    <div>
      <Header title="UI samples" />

      <Button>Click me!</Button>
      <Button variant="destructive">Click me!</Button>
      <Button variant="outline">Outline</Button>
      <Button variant="secondary">Secondary</Button>
      <Button variant="ghost">Ghost</Button>
      <Button variant="link">Click me!</Button>
      <Button variant="light">Light</Button>
      <div className="my-10">
        <Button loading={loading} onClick={hanleClick}>
          <Database />
          Loading
        </Button>
        <Button loading={loading} onClick={hanleClick} variant="secondary">
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
