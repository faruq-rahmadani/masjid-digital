import { zodResolver } from '@hookform/resolvers/zod'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { Navigate, useNavigate, useSearchParams } from 'react-router'
import { z } from 'zod'
import {
  Button,
  Callout,
  Checkbox,
  Field,
  Input,
  MosqueLogo,
  PasswordInput,
  Sheet,
  Stack,
} from '@/components'
import { usePublicMosqueLogo, usePublicMosqueProfile } from '@/hooks/useMosqueProfile'
import { useAuth } from '@/stores/sessionStore'

const schema = z.object({
  identifier: z.string().trim().min(1, 'Email atau nomor HP wajib diisi'),
  password: z.string().min(1, 'Password wajib diisi'),
  remember: z.boolean(),
})

/** Only same-app paths, so a crafted ?next= can't send people elsewhere after login. */
function safeNext(next: string | null) {
  return next?.startsWith('/') && !next.startsWith('//') ? next : '/'
}

// Login (E2-06). OTP WhatsApp from the mockup is P1 (E14-01) and left out for now.
export function LoginPage() {
  const { session, login } = useAuth()
  const mosque = usePublicMosqueProfile({ live: false })
  const logo = usePublicMosqueLogo({ live: false })
  const mosqueName = mosque.data?.name ?? 'Masjid Digital'
  const navigate = useNavigate()
  const [params] = useSearchParams()
  const [forgotOpen, setForgotOpen] = useState(false)
  const form = useForm({
    resolver: zodResolver(schema),
    defaultValues: { identifier: '', password: '', remember: true },
  })
  const { errors, isSubmitting } = form.formState
  const next = safeNext(params.get('next'))

  if (session) return <Navigate to={next} replace />

  const submit = form.handleSubmit(async (values) => {
    try {
      await login(values)
      navigate(next, { replace: true })
    } catch (error) {
      form.setError('root', {
        message: error instanceof Error ? error.message : 'Gagal masuk. Coba lagi.',
      })
    }
  })

  return (
    <div className="bg-brand -mx-[env(safe-area-inset-left)] flex min-h-dvh flex-col">
      <div className="pt-safe-10 text-on-brand flex flex-1 flex-col items-center justify-center gap-3.5 px-6 pb-10 text-center">
        {/* Public endpoints (no login). The placeholder mark and product name stand in while
            loading, and when the profile or logo hasn't been saved yet. */}
        {logo.data ? (
          <img
            src={logo.data}
            alt={`Logo ${mosqueName}`}
            className="bg-card size-18 rounded-full object-contain p-1.5"
          />
        ) : (
          <MosqueLogo />
        )}
        <h1 className="text-2xl font-extrabold">Panel Pengurus</h1>
        <p className="text-on-brand-muted text-base">{mosqueName}</p>
      </div>

      <form
        onSubmit={submit}
        noValidate
        aria-labelledby="login-title"
        className="rounded-t-sheet bg-canvas pb-safe-8 px-6 pt-7"
      >
        <Stack>
          <h2 id="login-title" className="text-xl font-extrabold">
            Masuk
          </h2>
          <Field label="Email atau nomor HP" error={errors.identifier?.message}>
            <Input
              size="lg"
              autoComplete="username"
              autoCapitalize="none"
              spellCheck={false}
              {...form.register('identifier')}
            />
          </Field>
          <Field label="Password" error={errors.password?.message}>
            <PasswordInput
              size="lg"
              autoComplete="current-password"
              {...form.register('password')}
            />
          </Field>
          <div className="flex items-center justify-between text-sm">
            <Checkbox label="Ingat saya" {...form.register('remember')} />
            <button
              type="button"
              onClick={() => setForgotOpen(true)}
              className="min-h-touch font-semibold underline underline-offset-2"
            >
              Lupa password?
            </button>
          </div>
          {errors.root?.message && (
            <Callout tone="danger" className="py-3">
              <span role="alert">{errors.root.message}</span>
            </Callout>
          )}
          <Button type="submit" size="xl" block loading={isSubmitting}>
            Masuk
          </Button>
          <p className="text-ink-muted text-center text-xs">
            Akun dibuat oleh super admin masjid. Belum punya akun? Hubungi pengurus.
          </p>
        </Stack>
      </form>

      {forgotOpen && (
        <Sheet
          open
          onClose={() => setForgotOpen(false)}
          title="Lupa password"
          footer={
            <Button size="xl" block onClick={() => setForgotOpen(false)}>
              Mengerti
            </Button>
          }
        >
          <p className="text-ink-muted">
            Minta super admin masjid mengatur ulang password Anda dari menu Pengguna. Setelah itu,
            masuk dengan password baru yang diberikan.
          </p>
        </Sheet>
      )}
    </div>
  )
}
