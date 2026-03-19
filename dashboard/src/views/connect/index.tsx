import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/shadcn/popover'
import { Typography } from '@/components/ui/typography'
import { useClusterConnect } from '@/lib/api/gen'
import { valibotResolver } from '@hookform/resolvers/valibot'
import { Database, EthernetPort, Info, KeyRound, Server, UserRound } from 'lucide-react'
import { Controller, useForm, type SubmitHandler } from 'react-hook-form'
import styles from './index.module.css'
import { validationSchema, type FormSubmitData } from './validation-schema'
import { toast } from 'sonner'
import { capitalize } from '@/lib/utils'
import { hasErrorField } from '@/lib/api/types'
import { useNavigate } from '@tanstack/react-router'
import { ErrorMessage } from '@/components/ui/error-message'
import {
  InputGroup,
  InputGroupAddon,
  InputGroupButton,
  InputGroupInput
} from '@/components/ui/shadcn/input-group'

export function ConnectView() {
  const navigate = useNavigate()

  const { isPending, mutate } = useClusterConnect({
    mutation: {
      onError: (error) => {
        toast.error(capitalize(error.message), {
          description: hasErrorField(error, 'reason') && error.reason,
          duration: 3000
        })
      },
      onSuccess: () => {
        navigate({
          to: '/'
        })
      }
    }
  })

  const { register, handleSubmit, control } = useForm({
    mode: 'onBlur',
    resolver: valibotResolver(validationSchema),
    defaultValues: {
      host: '',
      port: '',
      user: '',
      password: ''
    }
  })

  const onSubmit: SubmitHandler<FormSubmitData> = (values) => {
    if (values.database === '') {
      values.database = undefined
    }

    mutate({ data: { ...values } })
  }

  return (
    <div className="bg-section-box grid h-screen grid-cols-[2fr_1fr]">
      <form
        onSubmit={handleSubmit(onSubmit)}
        className="relative flex flex-col items-center justify-center"
      >
        <div className="w-full max-w-[700px] px-12 py-10">
          <div className="flex items-center justify-center">
            <div className="mr-2 size-14">
              <img src="/logo.svg" alt="App logo" />
            </div>
            <Typography variant="h1" as="h2">
              PgDashboard
            </Typography>
          </div>
          <div className="mb-10 pt-10">
            <ul>
              <li className="mb-5">
                <Label htmlFor="conncet-database-input">Database (optional)</Label>
                <InputGroup>
                  <InputGroupAddon>
                    <Database />
                  </InputGroupAddon>
                  <InputGroupInput
                    id="conncet-database-input"
                    placeholder="postgres"
                    className="pl-2!"
                    autoComplete="off"
                    {...register('database')}
                  />
                  <InputGroupAddon align="inline-end">
                    <Popover>
                      <PopoverTrigger asChild>
                        <InputGroupButton className="cursor-pointer rounded-full" size="icon-xs">
                          <Info />
                        </InputGroupButton>
                      </PopoverTrigger>
                      <PopoverContent className="text-sm leading-normal">
                        <Typography variant="small" as="h1">
                          By default dashboard tries to connect to <b>postgres</b> database
                        </Typography>
                      </PopoverContent>
                    </Popover>
                  </InputGroupAddon>
                </InputGroup>
              </li>
              <li className="flex">
                <div className="mr-8 w-full max-w-1/2">
                  <Controller
                    name="host"
                    control={control}
                    render={({ field, fieldState }) => (
                      <>
                        <Label htmlFor="connect-host-input">Host</Label>
                        <InputGroup className="mb-1">
                          <InputGroupInput
                            {...field}
                            id="connect-host-input"
                            placeholder="127.0.0.1"
                            autoComplete="off"
                            aria-invalid={fieldState.invalid}
                          />
                          <InputGroupAddon>
                            <Server />
                          </InputGroupAddon>
                        </InputGroup>
                        <ErrorMessage show={fieldState.invalid}>
                          {fieldState.error?.message}
                        </ErrorMessage>
                      </>
                    )}
                  />
                </div>
                <div>
                  <Controller
                    name="port"
                    control={control}
                    render={({ field, fieldState }) => (
                      <>
                        <Label htmlFor="conncet-port-input">Port</Label>
                        <InputGroup className="mb-1">
                          <InputGroupInput
                            {...field}
                            id="conncet-port-input"
                            type="number"
                            placeholder="5432"
                            autoComplete="off"
                            aria-invalid={fieldState.invalid}
                          />
                          <InputGroupAddon>
                            <EthernetPort />
                          </InputGroupAddon>
                        </InputGroup>
                        <ErrorMessage show={fieldState.invalid}>
                          {fieldState.error?.message}
                        </ErrorMessage>
                      </>
                    )}
                  />
                </div>
              </li>
              <li>
                <Controller
                  name="user"
                  control={control}
                  render={({ field, fieldState }) => (
                    <>
                      <Label htmlFor="conncet-user-input">User</Label>
                      <InputGroup className="mb-1">
                        <InputGroupInput
                          id="conncet-user-input"
                          placeholder="postgres"
                          autoComplete="off"
                          aria-invalid={fieldState.invalid}
                          {...field}
                        />
                        <InputGroupAddon>
                          <UserRound />
                        </InputGroupAddon>
                      </InputGroup>
                      <ErrorMessage show={fieldState.invalid}>
                        {fieldState.error?.message}
                      </ErrorMessage>
                    </>
                  )}
                />
              </li>
              <li>
                <Controller
                  name="password"
                  control={control}
                  render={({ field, fieldState }) => (
                    <>
                      <Label htmlFor="connect-password-input">Password</Label>
                      <InputGroup className="mb-1">
                        <InputGroupInput
                          id="connect-password-input"
                          type="password"
                          aria-invalid={fieldState.invalid}
                          autoComplete="off"
                          {...field}
                        />
                        <InputGroupAddon>
                          <KeyRound />
                        </InputGroupAddon>
                      </InputGroup>
                      <ErrorMessage show={fieldState.invalid}>
                        {fieldState.error?.message}
                      </ErrorMessage>
                    </>
                  )}
                />
              </li>
            </ul>
          </div>
          <Button type="submit" loading={isPending} fullWidth size="lg">
            Connect
          </Button>
        </div>

        <div className="absolute bottom-3">
          <Typography variant="muted" className="text-xs">
            PostgreSQL logo © PostgreSQL Global Development Group, used for illustrative purposes.
            This project is not affiliated with PostgreSQL.
          </Typography>
        </div>
      </form>

      <div className="flex items-center justify-center">
        <div className={styles.view} />
      </div>
    </div>
  )
}
