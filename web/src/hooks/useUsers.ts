import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as userService from '@/services/userService'
import type { CreateUserRequest, UpdateUserRequest, User } from '@/types/user'

// Separate list and detail keys, so refreshing the list doesn't also refetch every user detail.
const userKeys = {
  list: ['users', 'list'] as const,
  detail: (id: number) => ['users', 'detail', id] as const,
}

export function useUsers() {
  return useQuery({ queryKey: userKeys.list, queryFn: userService.listUsers })
}

export function useUser(id: number, { enabled = true } = {}) {
  return useQuery({
    queryKey: userKeys.detail(id),
    queryFn: () => userService.getUser(id),
    enabled,
  })
}

/**
 * POST /users answers only { message }, with no created user or id, so there is no row to insert:
 * the list is refetched. Once the API returns the new user, add it to the cache like useUpdateUser.
 */
export function useCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateUserRequest) => userService.createUser(body),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: userKeys.list }),
  })
}

/**
 * Sends the whole user with `changes` applied, since the API replaces every field (PUT). The API
 * returns the updated user, so only that row is replaced in the cache; nothing is refetched.
 */
export function useUpdateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: ({ user, changes }: { user: User; changes: Partial<UpdateUserRequest> }) =>
      userService.updateUser(user.id, {
        name: user.name,
        phone: user.phone,
        email: user.email,
        role: user.role,
        status: user.status,
        ...changes,
      }),
    onSuccess: (updated) => {
      queryClient.setQueryData<User[]>(userKeys.list, (list) =>
        list?.map((user) => (user.id === updated.id ? updated : user)),
      )
      queryClient.setQueryData(userKeys.detail(updated.id), updated)
    },
  })
}

/** Changes nothing the app shows, so no cache update. */
export function useResetPassword() {
  return useMutation({
    mutationFn: ({ id, password }: { id: number; password: string }) =>
      userService.resetUserPassword(id, password),
  })
}
