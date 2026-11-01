import { zodResolver } from '@hookform/resolvers/zod'
import { Controller, useForm } from 'react-hook-form'
import {
  ActionBar,
  Button,
  Callout,
  Card,
  Field,
  Input,
  PageBody,
  PageHeader,
  PhoneInput,
  Skeleton,
  Stack,
  Textarea,
  useToast,
} from '@/components'
import { useMosqueProfile, useUpdateMosqueProfile } from '@/hooks/useMosqueProfile'
import type { MosqueProfile } from '@/types/mosque'
import { normalizePhone } from '@/utils/format/phone'
import { LocationCard } from './components/LocationCard'
import { LogoCard } from './components/LogoCard'
import { profileSchema, toFormValues } from './schema'
import { guessTimezone } from './utils'

// Profil masjid (E3-03). Super admin and takmir; the route is guarded by RequireAccess.
export function ProfilePage() {
  const profile = useMosqueProfile()

  return (
    <>
      <PageHeader title="Profil Masjid" />
      {profile.isPending ? (
        <PageBody>
          <Card aria-busy aria-label="Memuat profil masjid" className="flex flex-col gap-3">
            <Skeleton className="h-11" />
            <Skeleton className="h-16" />
            <Skeleton className="h-11" />
          </Card>
        </PageBody>
      ) : profile.isError ? (
        <PageBody>
          <Callout
            tone="danger"
            title="Profil masjid gagal dimuat"
            action={
              <Button size="md" variant="secondary" onClick={() => profile.refetch()}>
                Coba lagi
              </Button>
            }
          >
            {profile.error.message}
          </Callout>
        </PageBody>
      ) : (
        // Mounted once the profile is known, so the form starts from the saved values.
        <ProfileForm profile={profile.data} />
      )}
    </>
  )
}

function ProfileForm({ profile }: { profile: MosqueProfile | null }) {
  const update = useUpdateMosqueProfile()
  const toast = useToast()
  const form = useForm({
    resolver: zodResolver(profileSchema),
    defaultValues: toFormValues(profile),
  })
  const { errors } = form.formState

  const submit = form.handleSubmit(({ location, contact_phone, ...values }) => {
    if (!location) return // the schema already reports it
    update.mutate(
      {
        ...values,
        contact_phone: contact_phone && (normalizePhone(contact_phone) ?? contact_phone),
        latitude: location.latitude,
        longitude: location.longitude,
      },
      {
        onSuccess: (saved) => {
          toast({ message: 'Profil masjid disimpan' })
          form.reset(toFormValues(saved))
        },
        onError: (error) => form.setError('root', { message: error.message }),
      },
    )
  })

  return (
    <>
      <PageBody>
        <LogoCard />
        <form id="mosque-profile" onSubmit={submit} noValidate className="flex flex-col gap-3.5">
          {!profile && (
            <Callout title="Profil belum diisi">
              Lengkapi nama dan lokasi masjid. Waktu sholat dihitung dari lokasi ini.
            </Callout>
          )}
          <Card>
            <Stack gap="sm">
              <Field label="Nama masjid" error={errors.name?.message}>
                <Input placeholder="Masjid Al-Ikhlas" {...form.register('name')} />
              </Field>
              <Field label="Alamat (opsional)" error={errors.address?.message}>
                <Textarea
                  placeholder="Jalan, RT/RW, kelurahan, kecamatan, kota"
                  {...form.register('address')}
                />
              </Field>
              <Field
                label="Kontak pengurus (opsional)"
                hint="Nomor HP atau WhatsApp"
                error={errors.contact_phone?.message}
              >
                <PhoneInput
                  placeholder="0812-3456-7890"
                  autoComplete="off"
                  {...form.register('contact_phone')}
                />
              </Field>
            </Stack>
          </Card>

          <Controller
            control={form.control}
            name="location"
            render={({ field, fieldState }) => (
              <LocationCard
                location={field.value}
                onLocationChange={(value) => {
                  field.onChange(value)
                  form.setValue('timezone', guessTimezone(value.longitude), { shouldDirty: true })
                }}
                locationError={fieldState.error?.message}
                timezone={form.watch('timezone')}
                onTimezoneChange={(value) =>
                  form.setValue('timezone', value, { shouldDirty: true })
                }
              />
            )}
          />

          {errors.root?.message && <Callout tone="danger">{errors.root.message}</Callout>}
        </form>
      </PageBody>

      <ActionBar>
        <Button type="submit" form="mosque-profile" size="xl" block loading={update.isPending}>
          Simpan profil
        </Button>
      </ActionBar>
    </>
  )
}
