import { SdkFunctionWrapper } from '@/graphql/generated'

export const useAPI = () => {
  const { $graphQLWithWrapper } = useNuxtApp()

  const wrapper: SdkFunctionWrapper = (action, _operationName, _operationType) => {
    return action()
  }

  return {
    '$graphql': $graphQLWithWrapper(wrapper),
  }
}