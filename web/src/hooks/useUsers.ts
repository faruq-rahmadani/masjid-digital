import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import * as userService from '@/services/userService'
import type { CreateUserRequest, UpdateUserRequest, User } from '@/types/user'

const userKeys = {
  list: ['users', 'list'] as const,
}

export function useUsers() {
  return useQuery({ queryKey: userKeys.list, queryFn: userService.listUsers })
}

export function useCreateUser() {
  const queryClient = useQueryClient()
  return useMutation({
    mutationFn: (body: CreateUserRequest) => userService.createUser(body),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: userKeys.list }),
  })
}

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
    },
  })
}

export function useResetPassword() {
  return useMutation({
    mutationFn: ({ id, password }: { id: number; password: string }) =>
      userService.resetUserPassword(id, password),
  })
}
